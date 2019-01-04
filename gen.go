package main

//go:generate go get github.com/roboll/go-vendorinstall
//go:generate go-vendorinstall -target ./toolbin github.com/ethereum/go-ethereum/cmd/abigen

//go:generate ./toolbin/abigen --abi ./blockchain/contracts/PaymentProcessor.abi --bin ./blockchain/contracts/PaymentProcessor.bin --out ./blockchain/contracts/PaymentProcessor.go --pkg contracts --type PaymentProcessorContract
