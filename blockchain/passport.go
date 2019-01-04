package blockchain

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/monetha/reputation-go-sdk/eth"
	"github.com/monetha/reputation-go-sdk/facts"
)

var factKey = []byte("ICO Data")
var factKeyBytes [32]byte

// WriteData writes data for the specific key
func WriteData(ctx context.Context, passport common.Address, ethClient *ethclient.Client, key *ecdsa.PrivateKey, factBytes []byte) (txHash common.Hash, err error) {
	ethSession := eth.New(ethClient, log.Warn)
	writeSession := ethSession.NewSession(key)
	provider := facts.NewProvider(writeSession)
	copy(factKeyBytes[:], factKey)
	txHash, err = provider.WriteTxData(ctx, passport, factKeyBytes, factBytes)
	return
}
