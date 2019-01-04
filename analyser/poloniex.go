package analyser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/monetha/ico-analyzer/types"
)

const poloniexURL = "https://poloniex.com/public?command=returnChartData&currencyPair=USDT_ETH&start=%d&end=%d&period=7200"

func getEthRateFromPoloneix(startDate, endDate int64) (startDateRate, endDateRate float64, err error) {
	resp, err := http.Get(fmt.Sprintf(poloniexURL, startDate, endDate))
	if err != nil {
		return
	}

	poloniexData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var ethRatesData []types.PoloniexData
	err = json.Unmarshal(poloniexData, &ethRatesData)
	if err != nil {
		return
	}

	startDateRate = ethRatesData[0].WeightedAverage
	endDateRate = ethRatesData[len(ethRatesData)-1].WeightedAverage
	return
}
