package config

import (
	"fmt"
	"os"
)

const (
	ethereumJSONRPCURLEnvName      = "ETHEREUM_JSON_RPC_URL"
	merchantKeyEnvName             = "MERCHANT_KEY"
	paymentProcessorAddressEnvName = "PAYMENT_PROCESSOR_ADDRESS"
)

var (
	// EthereumJSONRPCURL is to connected to ethereum client
	EthereumJSONRPCURL string
	//MerchantKey for processing payment and writing facts in passport address
	MerchantKey string
	//PaymentProcessorAddress for calling refundPayment and processPayment method
	PaymentProcessorAddress string
)

// Parse will parse all the flags into config variables
func Parse() error {
	var err error

	EthereumJSONRPCURL, err = getEnvString(ethereumJSONRPCURLEnvName)
	if err != nil {
		return err
	}

	MerchantKey, err = getEnvString(merchantKeyEnvName)
	if err != nil {
		return err
	}

	PaymentProcessorAddress, err = getEnvString(paymentProcessorAddressEnvName)
	return err
}

func getEnvString(envName string) (string, error) {
	if value, ok := os.LookupEnv(envName); ok {
		return value, nil
	}

	return "", fmt.Errorf("environment variable %v not found", envName)
}
