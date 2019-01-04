package types

const (
	// ReceiptStatusFailed is the status code of a transaction if execution failed.
	ReceiptStatusFailed = uint64(0)
	// ReceiptStatusSuccessful is the status code of a transaction if execution succeeded.
	ReceiptStatusSuccessful = uint64(1)
	// OrderStateNull is order state for null order
	OrderStateNull = uint8(0)
	// OrderStatePaid is order state for paid order
	OrderStatePaid = uint8(2)
)

// ICORatingData stores data fetched from ico rating website
type ICORatingData struct {
	CfrCurrency      string  `json:"cfr_currency"`
	Cfr              float64 `json:"cfr"`
	IcoStartDate     string  `json:"ico_start_date"`
	IcoEndDate       string  `json:"ico_end_date"`
	IcoPriceCur      string  `json:"ico_price_cur"`
	IcoPrice         float64 `json:"ico_price"`
	IcoPriceAdjusted float64 `json:"ico_price_adjusted"`
}

// EtherScanAllTxns stores data fetched from etherscan for all token txns
type EtherScanAllTxns struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Result  []Result `json:"result"`
}

// EtherScanTrxWithErr stores data fetched from etherscan for all crowsale txns
type EtherScanTrxWithErr struct {
	Status  string              `json:"status"`
	Message string              `json:"message"`
	Result  []TransactionResult `json:"result"`
}

// EtherScanBalance stores data fetched from etherscan for balance of a specific address
type EtherScanBalance struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// EtherScanIntTxns stores data fetched from etherscan for all inetnal txns
type EtherScanIntTxns struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Result  []IntTxnResult `json:"result"`
}

// Result stores transaction data result
type Result struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      string `json:"tokenDecimal"`
	TransactionIndex  string `json:"transactionIndex"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Input             string `json:"input"`
	Confirmations     string `json:"confirmations"`
}

// TransactionResult stores transaction data result
type TransactionResult struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	TransactionIndex  string `json:"transactionIndex"`
	From              string `json:"from"`
	To                string `json:"to"`
	Value             string `json:"value"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	IsError           string `json:"isError"`
	TxreceiptStatus   string `json:"txreceipt_status"`
	Input             string `json:"input"`
	ContractAddress   string `json:"contractAddress"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	GasUsed           string `json:"gasUsed"`
	Confirmations     string `json:"confirmations"`
}

// IntTxnResult stores inetrnalTx data from etherscan of a specific address
type IntTxnResult struct {
	BlockNumber     string `json:"blockNumber"`
	TimeStamp       string `json:"timeStamp"`
	Hash            string `json:"hash"`
	From            string `json:"from"`
	To              string `json:"to"`
	Value           string `json:"value"`
	ContractAddress string `json:"contractAddress"`
	Input           string `json:"input"`
	Type            string `json:"type"`
	Gas             string `json:"gas"`
	GasUsed         string `json:"gasUsed"`
	TraceID         string `json:"traceId"`
	IsError         string `json:"isError"`
	ErrCode         string `json:"errCode"`
}

// PoloniexData stores data fetched from poloniex for ether rates
type PoloniexData struct {
	Date            int     `json:"date"`
	High            float64 `json:"high"`
	Low             float64 `json:"low"`
	Open            float64 `json:"open"`
	Close           float64 `json:"close"`
	Volume          float64 `json:"volume"`
	QuoteVolume     float64 `json:"quoteVolume"`
	WeightedAverage float64 `json:"weightedAverage"`
}

// CalculatedData stores final analysed data for an ICO
type CalculatedData struct {
	TokensIssued     int64   `json:"tokens_issued"`
	EfrToken         float64 `json:"efr_token"`
	EthRateStart     float64 `json:"eth_rate_start"`
	EthRateEnd       float64 `json:"eth_rate_end"`
	EfrTokenAdjusted float64 `json:"efr_token_adjusted"`
	TokenCheckResult struct {
		FundsRaisedDiff         float64 `json:"funds_raised_diff"`
		FundsRaisedAdjustedDiff float64 `json:"funds_raised_adjusted_diff"`
		FundsRaisedCheck        string  `json:"funds_raised_check"`
	} `json:"token_check_result"`
	IcoEthIn             float64 `json:"ico_eth_in"`
	IcoEthOut            int64   `json:"ico_eth_out"`
	IcoEthTotal          int64   `json:"ico_eth_total"`
	EfrIcoTx             float64 `json:"efr_ico_tx"`
	EfrOwnerTxCurrency   string  `json:"efr_owner_tx_currency"`
	IcoWalletCheckResult struct {
		FundsRaisedDiff  float64 `json:"funds_raised_diff"`
		FundsRaisedCheck string  `json:"funds_raised_check"`
	} `json:"ico_wallet_check_result"`
	Metrics struct {
		DistributionDays              float64 `json:"distribution_days"`
		DistributionStartFromIcoStart string  `json:"distribution_start_from_ico_start"`
		DistributionEndFromIcoEnd     string  `json:"distribution_end_from_ico_end"`
		FundsBalanceEth               float64 `json:"funds_balance_eth"`
	} `json:"metrics"`
}

// ICOAnalyzerData is data recieved from web app for ico analysis
type ICOAnalyzerData struct {
	Version                int     `json:"version"`
	IcoName                string  `json:"icoName"`
	Decimals               int     `json:"decimals"`
	TokenContractAddress   string  `json:"tokenContractAddress"`
	CrowdSaleAddress       string  `json:"crowdsaleAddress"`
	OwnerAddress           string  `json:"ownerAddress"`
	TokenIssuerAddress     string  `json:"tokenIssuerAddress"`
	FundAddress            string  `json:"fundAddress"`
	EthNominated           bool    `json:"ethNominated"`
	TokenTxInputAdjustment bool    `json:"token_tx_input_adjustment"`
	OwnerIsIcoWallet       bool    `json:"owner_is_ico_wallet"`
	Confidence             float64 `json:"confidence"`
	PassportAddress        string  `json:"passportAddress"`
	TxHash                 string  `json:"txHash"`
	OrderID                int64   `json:"orderId"`
	AccountAddress         string  `json:"accountAddress"`
}

// ICOPassport contains complete ico passport data
type ICOPassport struct {
	Metadata       ICOAnalyzerData `json:"metadata"`
	IcoInfo        ICORatingData   `json:"ico_info"`
	CalculatedData CalculatedData  `json:"calculated_data"`
}
