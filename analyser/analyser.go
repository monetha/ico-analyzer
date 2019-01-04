package analyser

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/monetha/ico-analyzer/types"
)

// Run will run the analyser
func Run(ctx context.Context, data *types.ICOPassport) (analysedData types.CalculatedData, icoRatingData types.ICORatingData, err error) {
	if data.Metadata.Version != 0 {
		icoRatingData = data.IcoInfo
		analysedData = data.CalculatedData

		timeLayout := "02 Jan 2006"
		icoStartDate, err := time.Parse(timeLayout, data.IcoInfo.IcoStartDate)
		if err != nil {
			return analysedData, icoRatingData, err
		}

		icoEndDate, err := time.Parse(timeLayout, data.IcoInfo.IcoEndDate)
		if err != nil {
			return analysedData, icoRatingData, err
		}
		startDateEthRate, endDateEthRate, err := getEthRateFromPoloneix(icoStartDate.Unix(), icoEndDate.Unix())
		if err != nil {
			return analysedData, icoRatingData, err
		}
		icoRatingData.IcoPriceAdjusted = icoRatingData.IcoPrice * (endDateEthRate / startDateEthRate)
		analysedData.EthRateStart = startDateEthRate
		analysedData.EthRateEnd = endDateEthRate
		ethPriceFluctuation := math.Abs((endDateEthRate - startDateEthRate) / startDateEthRate)
		analysedData.EfrToken = float64(analysedData.TokensIssued) * icoRatingData.IcoPrice
		analysedData.EfrTokenAdjusted = float64(analysedData.TokensIssued) * icoRatingData.IcoPriceAdjusted
		analysedData.TokenCheckResult.FundsRaisedDiff = (analysedData.EfrToken - icoRatingData.Cfr) / icoRatingData.Cfr
		analysedData.TokenCheckResult.FundsRaisedAdjustedDiff = (analysedData.EfrTokenAdjusted - icoRatingData.Cfr) / icoRatingData.Cfr

		ethPriceFluctuation = math.Max(ethPriceFluctuation, data.Metadata.Confidence)
		analysedData.TokenCheckResult.FundsRaisedCheck = "Passed"
		if analysedData.TokenCheckResult.FundsRaisedDiff < 0 && (math.Abs(analysedData.TokenCheckResult.FundsRaisedDiff) > ethPriceFluctuation) {
			analysedData.TokenCheckResult.FundsRaisedCheck = "Failed"
		}

		data.Metadata.FundAddress = data.Metadata.OwnerAddress
		data.Metadata.OwnerIsIcoWallet = true
		if data.Metadata.CrowdSaleAddress != "" && data.Metadata.OwnerAddress != data.Metadata.CrowdSaleAddress {
			data.Metadata.FundAddress = data.Metadata.CrowdSaleAddress
			data.Metadata.OwnerIsIcoWallet = false
			var crowdSaleBalance float64
			var txnCount int64
			var ethBalance float64
			var fundAddress string

			crowdSaleBalance, txnCount, err = getCrowdSaleBalance(strings.ToLower(data.Metadata.FundAddress))
			if err != nil {
				return analysedData, icoRatingData, err
			}

			fundAddress, ethBalance, err = getEthBalance(strings.ToLower(data.Metadata.FundAddress))
			if err != nil {
				return analysedData, icoRatingData, err
			}

			endDateEthRate = math.Max(endDateEthRate, startDateEthRate)
			data.Metadata.FundAddress = fundAddress
			analysedData.Metrics.FundsBalanceEth = ethBalance
			analysedData.IcoEthIn = crowdSaleBalance
			analysedData.IcoEthOut = 0
			analysedData.IcoEthTotal = txnCount
			analysedData.EfrIcoTx = crowdSaleBalance * endDateEthRate
		}

		analysedData.EfrOwnerTxCurrency = icoRatingData.CfrCurrency
		analysedData.IcoWalletCheckResult.FundsRaisedDiff = (analysedData.EfrIcoTx - float64(icoRatingData.Cfr)) / float64(icoRatingData.Cfr)
		analysedData.IcoWalletCheckResult.FundsRaisedCheck = "Passed"

		if analysedData.IcoWalletCheckResult.FundsRaisedDiff < 0 && (math.Abs(analysedData.IcoWalletCheckResult.FundsRaisedDiff) > ethPriceFluctuation) {
			analysedData.IcoWalletCheckResult.FundsRaisedCheck = "Failed"
		}

		analysedData.Metrics.DistributionDays = icoEndDate.Sub(icoStartDate).Hours() / 24
		analysedData.Metrics.DistributionStartFromIcoStart = data.CalculatedData.Metrics.DistributionStartFromIcoStart
		analysedData.Metrics.DistributionEndFromIcoEnd = data.CalculatedData.Metrics.DistributionEndFromIcoEnd
		return analysedData, icoRatingData, nil
	}

	icoRatingData, icoStartDate, icoEndDate, err := icoRating(data.Metadata.IcoName)
	if err != nil {
		return analysedData, icoRatingData, err
	}

	totalSupply, tokenIssuingAddress, tokenStartDate, tokenEndDate, err := getTokenCount(strings.ToLower(data.Metadata.TokenContractAddress), data.Metadata.Decimals, icoEndDate)
	if err != nil {
		return analysedData, icoRatingData, err
	}
	data.Metadata.TokenIssuerAddress = tokenIssuingAddress
	data.Metadata.Confidence = 0.1

	startDateEthRate, endDateEthRate, err := getEthRateFromPoloneix(icoStartDate.Unix(), icoEndDate.Unix())
	if err != nil {
		return analysedData, icoRatingData, err
	}
	icoRatingData.IcoPriceAdjusted = icoRatingData.IcoPrice * (endDateEthRate / startDateEthRate)
	analysedData.EthRateStart = startDateEthRate
	analysedData.EthRateEnd = endDateEthRate
	ethPriceFluctuation := math.Abs((endDateEthRate - startDateEthRate) / startDateEthRate)
	analysedData.TokensIssued = int64(totalSupply)
	analysedData.EfrToken = float64(analysedData.TokensIssued) * icoRatingData.IcoPrice
	analysedData.EfrTokenAdjusted = float64(analysedData.TokensIssued) * icoRatingData.IcoPriceAdjusted
	analysedData.TokenCheckResult.FundsRaisedDiff = (analysedData.EfrToken - icoRatingData.Cfr) / icoRatingData.Cfr
	analysedData.TokenCheckResult.FundsRaisedAdjustedDiff = (analysedData.EfrTokenAdjusted - icoRatingData.Cfr) / icoRatingData.Cfr

	ethPriceFluctuation = math.Max(ethPriceFluctuation, data.Metadata.Confidence)
	analysedData.TokenCheckResult.FundsRaisedCheck = "Passed"
	if analysedData.TokenCheckResult.FundsRaisedDiff < 0 && (math.Abs(analysedData.TokenCheckResult.FundsRaisedDiff) > ethPriceFluctuation) {
		analysedData.TokenCheckResult.FundsRaisedCheck = "Failed"
	}

	data.Metadata.FundAddress = data.Metadata.OwnerAddress
	data.Metadata.OwnerIsIcoWallet = true
	if data.Metadata.CrowdSaleAddress != "" && data.Metadata.OwnerAddress != data.Metadata.CrowdSaleAddress {
		data.Metadata.FundAddress = data.Metadata.CrowdSaleAddress
		data.Metadata.OwnerIsIcoWallet = false
	}
	data.Metadata.EthNominated = true
	data.Metadata.TokenTxInputAdjustment = false
	var crowdSaleBalance float64
	var ethBalance float64
	var txnCount int64
	var fundAddress string

	if data.Metadata.FundAddress != "" {
		crowdSaleBalance, txnCount, err = getCrowdSaleBalance(strings.ToLower(data.Metadata.FundAddress))
		if err != nil {
			return analysedData, icoRatingData, err
		}
	}

	fundAddress, ethBalance, err = getEthBalance(strings.ToLower(data.Metadata.FundAddress))
	if err != nil {
		return analysedData, icoRatingData, err
	}
	data.Metadata.FundAddress = fundAddress
	analysedData.Metrics.FundsBalanceEth = ethBalance

	endDateEthRate = math.Max(endDateEthRate, startDateEthRate)
	analysedData.IcoEthIn = crowdSaleBalance
	analysedData.IcoEthOut = 0
	analysedData.IcoEthTotal = txnCount
	analysedData.EfrIcoTx = crowdSaleBalance * endDateEthRate
	analysedData.EfrOwnerTxCurrency = icoRatingData.CfrCurrency
	analysedData.IcoWalletCheckResult.FundsRaisedDiff = (analysedData.EfrIcoTx - float64(icoRatingData.Cfr)) / float64(icoRatingData.Cfr)
	analysedData.IcoWalletCheckResult.FundsRaisedCheck = "Passed"

	if analysedData.IcoWalletCheckResult.FundsRaisedDiff < 0 && (math.Abs(analysedData.IcoWalletCheckResult.FundsRaisedDiff) > ethPriceFluctuation) {
		analysedData.IcoWalletCheckResult.FundsRaisedCheck = "Failed"
	}

	analysedData.Metrics.DistributionDays = icoEndDate.Sub(icoStartDate).Hours() / 24
	analysedData.Metrics.DistributionStartFromIcoStart = tokenStartDate
	analysedData.Metrics.DistributionEndFromIcoEnd = tokenEndDate
	return analysedData, icoRatingData, nil
}
