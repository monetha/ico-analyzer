package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/chequebook"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/monetha/ico-analyzer/analyser"
	"github.com/monetha/ico-analyzer/blockchain"
	"github.com/monetha/ico-analyzer/blockchain/contracts"
	"github.com/monetha/ico-analyzer/config"
	"github.com/monetha/ico-analyzer/types"
)

// ProcessingGasLimit a maximum gas limit to be used for payment processing operation
const ProcessingGasLimit uint64 = 100000

func init() {
	path, found := os.LookupEnv("SSM_PS_PATH")
	if found {
		sess, err := session.NewSession(aws.NewConfig())
		if err != nil {
			log.Printf("error: failed to create new session %s\n", err)
			os.Exit(1)
		}
		service := ssm.New(sess)
		withDecryption := true
		request := ssm.GetParametersByPathInput{Path: &path, WithDecryption: &withDecryption}
		response, err := service.GetParametersByPath(&request)
		if err != nil {
			log.Printf("error: failed to get parameters, %s\n", err)
			os.Exit(1)
		}
		for _, parameter := range response.Parameters {
			paramName := strings.TrimPrefix(*parameter.Name, path)
			if err := os.Setenv(paramName, *parameter.Value); err != nil {
				log.Printf("error: failed to set environment variable %v: %v\n", paramName, err)
				os.Exit(1)
			}
			log.Printf("set env variable: %s\n", paramName)
		}
	}
}

func main() {
	lambda.Start(router)
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		return get(req)
	case "POST":
		return post(req)
	case "OPTIONS":
		return options(req)

	default:
		return clientError(http.StatusMethodNotAllowed)
	}
}

func options(_ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Origin, X-Requested-With, Content-Type, Accept",
		},
		Body: http.StatusText(200),
	}, nil
}

func get(_ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string("Hello123 world!"),
	}, nil
}

func post(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.Headers["Content-Type"] != "application/json" {
		return clientError(http.StatusNotAcceptable)
	}
	err := config.Parse()
	if err != nil {
		log.Printf("error: failed to parse config: %v", err)
		return clientError(http.StatusInternalServerError)
	}

	data := new(types.ICOPassport)
	err = json.Unmarshal([]byte(req.Body), data)
	if err != nil {
		log.Printf("error: failed to unmarshal request body: %v", err)
		return clientError(http.StatusUnprocessableEntity)
	}

	ethClient, err := ethclient.Dial(config.EthereumJSONRPCURL)
	if err != nil {
		log.Printf("error: failed to dial JSON-RPC (%v): %v", config.EthereumJSONRPCURL, err)
		return clientError(http.StatusInternalServerError)
	}
	privateKey, err := crypto.HexToECDSA(config.MerchantKey)
	if err != nil {
		log.Printf("error: failed to parse ECDSA private key from the given key: %v", err)
		return clientError(http.StatusInternalServerError)
	}

	err = runAnalyser(context.Background(), *data, ethClient, privateKey)
	if err != nil {
		return clientError(http.StatusInternalServerError)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
		Body: fmt.Sprintf("ICO analysis successfully done for token address : %s for which payment is done by txn : %s", data.Metadata.TokenContractAddress, data.Metadata.TxHash),
	}, nil
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
		Body: http.StatusText(status),
	}, nil
}

func runAnalyser(ctx context.Context, data types.ICOPassport, ethClient *ethclient.Client, privateKey *ecdsa.PrivateKey) error {

	transactOpts := bind.NewKeyedTransactor(privateKey)

	paymentProcessor, err := contracts.NewPaymentProcessorContract(common.HexToAddress(config.PaymentProcessorAddress), ethClient)
	if err != nil {
		log.Printf("error: failed to create an instance of payment processor contract: %v", err)
		return err
	}

	if err := checkTx(ctx, ethClient, common.HexToHash(data.Metadata.TxHash), data.Metadata.OrderID, paymentProcessor); err != nil {
		log.Printf("error: transaction %s processing failed: %v", data.Metadata.TxHash, err)
		return err
	}

	transactOpts.GasLimit = ProcessingGasLimit
	analysedData, icoRatingData, err := analyser.Run(ctx, &data)
	if err != nil {
		log.Printf("error: analyser failed calling RefundPayment for orderId %d: %v", data.Metadata.OrderID, err)

		txn, err := paymentProcessor.RefundPayment(transactOpts, big.NewInt(data.Metadata.OrderID), 0, 0, big.NewInt(0), "error")
		if err != nil {
			log.Printf("error: calling refund payment failed for orderId %d: %v", data.Metadata.OrderID, err)
			return err
		}

		if err = waitForTx(ctx, ethClient, txn.Hash()); err != nil {
			log.Printf("error: refund payment transaction %s failed for orderId %d: %v", txn.Hash(), data.Metadata.OrderID, err)
			return err
		}

		txn, err = paymentProcessor.WithdrawRefund(transactOpts, big.NewInt(data.Metadata.OrderID))
		if err != nil {
			log.Printf("error: calling refund payment failed for orderId %d: %v", data.Metadata.OrderID, err)
			return err
		}

		if err = waitForTx(ctx, ethClient, txn.Hash()); err != nil {
			log.Printf("error: withdraw refund payment transaction %s failed for orderId %d: %v", txn.Hash(), data.Metadata.OrderID, err)
		}
		return err
	}

	icoPassport := getICOPassport(analysedData, icoRatingData, data)
	icoPassportBytes, err := json.Marshal(icoPassport)
	if err != nil {
		log.Printf("error: marshaling icoPassport data into JSON failed: %v", err)
		return err
	}

	fmt.Println(string(icoPassportBytes))

	txHash, err := blockchain.WriteData(ctx, common.HexToAddress(data.Metadata.PassportAddress), ethClient, privateKey, icoPassportBytes)
	if err != nil {
		log.Printf("error: writing data on passport %s failed: %v", data.Metadata.PassportAddress, err)
		return err
	}

	if err = waitForTx(ctx, ethClient, txHash); err != nil {
		log.Printf("error: write to passport transaction %s failed: %v", txHash, err)
		return err
	}

	txn, err := paymentProcessor.ProcessPayment(transactOpts, big.NewInt(data.Metadata.OrderID), 0, 0, big.NewInt(0))
	if err != nil {
		log.Printf("error: calling process payment failed for orderId %d: %v", data.Metadata.OrderID, err)
		return err
	}

	if err = waitForTx(ctx, ethClient, txn.Hash()); err != nil {
		log.Printf("error: process payment transaction %s failed for orderId %d: %v", txn.Hash(), data.Metadata.OrderID, err)
		return err
	}

	return nil
}

func checkTx(ctx context.Context, backend chequebook.Backend, txHash common.Hash, orderID int64, paymentProcessor *contracts.PaymentProcessorContract) (err error) {
	if err = waitForTx(ctx, backend, txHash); err != nil {
		return // Transaction Failed
	}
	order, err := paymentProcessor.Orders(nil, big.NewInt(orderID))
	if err != nil {
		return
	}

	if order.State == types.OrderStateNull {
		return fmt.Errorf("order with order id : %d does not exist", orderID)
	}

	if order.State != types.OrderStatePaid {
		return fmt.Errorf("order with order id : %d is not in paid state", orderID)
	}
	return
}

func waitForTx(ctx context.Context, backend chequebook.Backend, txHash common.Hash) error {
	log.Printf("Waiting for transaction: 0x%x", txHash)

	type commiter interface {
		Commit()
	}
	if sim, ok := backend.(commiter); ok {
		sim.Commit()
		tr, err := backend.TransactionReceipt(ctx, txHash)
		if err != nil {
			return err
		}
		if tr.Status != types.ReceiptStatusSuccessful {
			return fmt.Errorf("tx failed: %+v", tr)
		}
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(4 * time.Second):
		}

		tr, err := backend.TransactionReceipt(ctx, txHash)
		if err != nil {
			if err == ethereum.NotFound {
				continue
			} else {
				return err
			}
		} else {
			if tr.Status != types.ReceiptStatusSuccessful {
				return fmt.Errorf("tx failed: %+v", tr)
			}
			return nil
		}
	}
}

func getICOPassport(analysedData types.CalculatedData, icoData types.ICORatingData, icoAnalyserData types.ICOPassport) types.ICOPassport {
	return types.ICOPassport{
		IcoInfo:        icoData,
		CalculatedData: analysedData,
		Metadata:       icoAnalyserData.Metadata,
	}
}
