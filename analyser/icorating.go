package analyser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/monetha/ico-analyzer/types"
	"github.com/moovweb/gokogiri"
)

const baseURL = "https://icorating.com/ico/%s/"

func icoRating(icoName string) (data types.ICORatingData, icoStartDate time.Time, icoEndDate time.Time, err error) {
	resp, err := http.Get(fmt.Sprintf(baseURL, icoName))
	if err != nil {
		return
	}

	icoInfoRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	doc, err := gokogiri.ParseHtml(icoInfoRaw)
	if err != nil {
		return
	}

	tables, err := doc.Search("//table[contains(@class, 'c-info-table--va-top')]")
	if err != nil {
		return
	}

	if len(tables) == 0 {
		err := errors.New("Invalid Ico name")
		return data, icoStartDate, icoEndDate, err
	}

	tableRows, err := tables[0].Search("//tr")
	if err != nil {
		return
	}

	dataMap := make(map[string]string, len(tableRows))
	for _, row := range tableRows {
		s1 := strings.Split(strings.TrimSpace(row.Content()), "\n")
		dataMap[s1[0]] = strings.TrimSpace(s1[len(s1)-1])
	}

	raised := strings.Split(dataMap["Raised"], " ")
	claimedFundsRaised, err := strconv.Atoi(strings.Replace(raised[0], ",", "", 10))
	if err != nil {
		return
	}
	claimedFundsRaisedCurreny := raised[1]

	price := strings.Split(strings.TrimPrefix(dataMap["Price"], "= "), " ")
	icoPrice, err := strconv.ParseFloat(price[0], 64)
	if err != nil {
		return
	}
	icoPriceCurrency := price[1]

	timeLayout := "02 Jan 2006"
	icoStartDate, err = time.Parse(timeLayout, dataMap["ICO start date"])
	if err != nil {
		return
	}

	icoEndDate, err = time.Parse(timeLayout, dataMap["ICO end date"])
	if err != nil {
		return
	}

	data = types.ICORatingData{
		Cfr:          float64(claimedFundsRaised),
		CfrCurrency:  claimedFundsRaisedCurreny,
		IcoStartDate: icoStartDate.Format(timeLayout),
		IcoEndDate:   icoEndDate.Format(timeLayout),
		IcoPrice:     icoPrice,
		IcoPriceCur:  icoPriceCurrency,
	}
	return
}
