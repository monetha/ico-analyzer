package analyser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/monetha/ico-analyzer/types"
)

const (
	noTxnFoundMsg                      = "No transactions found"
	etherScanBalance                   = "https://api.etherscan.io/api?module=account&action=balance&address=%s&tag=latest&apikey=YourApiKeyToken"
	etherScanURLForExternalTxns        = "https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&page=%d&offset=10000&sort=asc"
	etherScanURLForInternalTxns        = "https://api.etherscan.io/api?module=account&action=txlistinternal&address=%s&startblock=0&endblock=99999999&page=%d&offset=10000&sort=asc&apikey=YourApiKeyToken"
	etherScanURLForFund                = "https://api.etherscan.io/api?module=account&action=txlistinternal&address=%s&startblock=0&endblock=99999999&page=%d&offset=200&sort=asc&apikey=YourApiKeyToken"
	etherScanURLForTokenCount          = "https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=%s&address=%s&page=%d&offset=10000&sort=asc&apikey=YourApiKeyToken"
	etherScanURLForTokenIssuingAddress = "https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=%s&page=%d&offset=200&sort=asc&apikey=YourApiKeyToken"
	maxOffset                          = 10000
)

func getCrowdSaleBalance(address string) (balance float64, txnCount int64, err error) {

	extBalance, extTxCount, err := getBalance(etherScanURLForExternalTxns, address)
	if err != nil {
		return
	}

	internalBalance, inetrnalTxCount, err := getBalance(etherScanURLForInternalTxns, address)
	if err != nil {
		return
	}

	balance = extBalance + internalBalance
	txnCount = extTxCount + inetrnalTxCount

	return
}

func getTokenCount(tokenAddress string, tokenDecimals int, icoEndDate time.Time) (tokenCount float64, tokenIssuingAddress string, tokenStartDate string, tokenEndDate string, err error) {
	resp, err := http.Get(fmt.Sprintf(etherScanURLForTokenIssuingAddress, tokenAddress, 1))
	if err != nil {
		return
	}

	txnInfoRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var txnData types.EtherScanAllTxns
	err = json.Unmarshal(txnInfoRaw, &txnData)
	if err != nil {
		return
	}

	tokenIssuingAddress = maxOccurrence(txnData.Result)

	var page = 1
	icoEndDateEpoch := icoEndDate.Unix()
	tokenStartDateTmp := int64(0)
	temp := int64(0)
	// FLAGS
	isTokenStrtDateSet := false
	tokenEndDate = "" //just to be sure that nothing is passed in to the function.

	for {
		resp, err := http.Get(fmt.Sprintf(etherScanURLForTokenCount, tokenAddress, tokenIssuingAddress, page))
		if err != nil {
			return 0, "", "", "", err
		}

		txnInfoRaw, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, "", "", "", err
		}

		var txnData types.EtherScanAllTxns
		err = json.Unmarshal(txnInfoRaw, &txnData)
		if err != nil {
			return 0, "", "", "", err
		}

		if txnData.Message == noTxnFoundMsg {
			break
		}

		for _, txn := range txnData.Result {
			txnValue, err := strconv.ParseFloat(txn.Value, 64)
			if err != nil {
				return 0, "", "", "", err
			}
			if !isTokenStrtDateSet && txnValue != 0 { //checks for the first non zero txnValue and sets tokenStartDate
				tokenStartDateTmp, err = strconv.ParseInt(txn.TimeStamp, 10, 64)
				if err != nil {
					return 0, "", "", "", err
				}
				tokenStartDateTime := time.Unix(tokenStartDateTmp, 0)
				timeLayout := "02 Jan 2006"
				tokenStartDate = tokenStartDateTime.Format(timeLayout)
				isTokenStrtDateSet = true
			} else {
				tokenStartDateTmp, err = strconv.ParseInt(txn.TimeStamp, 10, 64)
				if err != nil {
					return 0, "", "", "", err
				}
				if tokenStartDateTmp > icoEndDateEpoch && tokenEndDate == "" {
					if temp == 0 {
						temp = icoEndDateEpoch
					}
					tokenEndDateTime := time.Unix(temp, 0)
					timeLayout := "02 Jan 2006"
					tokenEndDate = tokenEndDateTime.Format(timeLayout)

				}
				temp = tokenStartDateTmp

			}

			tokenCount += txnValue / math.Pow10(tokenDecimals)
		}

		if len(txnData.Result) < maxOffset {
			break
		}
		page++
	}

	if tokenEndDate == "" {
		tokenEndDateTime := time.Unix(icoEndDateEpoch, 0)
		timeLayout := "02 Jan 2006"
		tokenEndDate = tokenEndDateTime.Format(timeLayout)
	}
	return
}

func maxOccurrence(data []types.Result) (address string) {
	addressCount := make(map[string]int64, 200)
	var max int64
	for _, d := range data {
		addressCount[d.From]++
		if addressCount[d.From] >= max {
			max = addressCount[d.From]
			address = d.From
		}
	}
	return
}

func maxOccurrenceFund(data []types.IntTxnResult) (address string) {
	addressCount := make(map[string]int64, 200)
	var max int64
	for _, d := range data {
		addressCount[d.To]++
		if addressCount[d.To] >= max {
			max = addressCount[d.To]
			address = d.To
		}
	}
	return
}
func getBalance(url string, address string) (balance float64, txnCount int64, err error) {
	var page = 1
	txnCount = 0
	for {
		resp, err := http.Get(fmt.Sprintf(url, address, page))
		if err != nil {
			return balance, txnCount, err
		}

		txnInfoRaw, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return balance, txnCount, err
		}

		var txnData types.EtherScanTrxWithErr
		err = json.Unmarshal(txnInfoRaw, &txnData)
		if err != nil {
			return balance, txnCount, err
		}

		if txnData.Message == noTxnFoundMsg {
			break
		}

		for _, txn := range txnData.Result {
			if txn.IsError != "1" && txn.To == address {
				txnValue, err := strconv.ParseFloat(txn.Value, 64)
				if err != nil {
					return balance, txnCount, err
				}
				balance += txnValue / math.Pow10(18)
				txnCount++
			}
		}

		if len(txnData.Result) < maxOffset {
			break
		}
		page++
	}
	return
}

func getEthBalance(address string) (fundAddress string, ethBalance float64, err error) {
	resp, err := http.Get(fmt.Sprintf(etherScanURLForFund, address, 1))
	if err != nil {
		return
	}

	txnInfoRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var txnDataNew types.EtherScanIntTxns
	err = json.Unmarshal(txnInfoRaw, &txnDataNew)
	if err != nil {
		return
	}

	fundAddress = maxOccurrenceFund(txnDataNew.Result)

	if fundAddress == "" {
		fundAddress = address
	}

	resp, err = http.Get(fmt.Sprintf(etherScanBalance, fundAddress))
	if err != nil {
		return
	}

	txnInfoRaw, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var balanceData types.EtherScanBalance
	err = json.Unmarshal(txnInfoRaw, &balanceData)
	if err != nil {
		return
	}

	ethBalance, err = strconv.ParseFloat(balanceData.Result, 64)
	if err != nil {
		return
	}
	ethBalance = ethBalance / math.Pow10(18)

	return
}
