package oms

import (
	"context"
	"directionalMaker/data"
	"github.com/adshao/go-binance/v2/futures"
	"log"
	"strconv"
)

func convertPriceToString(price float64, settings data.SymbolSettings) string {

	return ""
}

func BinanceSide(side Side) futures.SideType {
	switch side {
	case SideLong:
		return futures.SideTypeBuy
	case SideShort:
		return futures.SideTypeSell
	case SideNone:
		log.Println("BinanceSide conversion ERROR! SideNone passed, why?!")
		return ""
	default:
		return ""
	}
}

func BinanceSideToInternal(binanceSide futures.SideType) Side {
	switch binanceSide {
	case futures.SideTypeBuy:
		return SideLong
	case futures.SideTypeSell:
		return SideShort
	default:
		log.Println("BinanceSideToInternal: something went wrong, empty futures.SideType passed!")
		return SideNone
	}
}

func BinancePrice(priceFloat float64, tickSize string) string {
	// Calculate number of decimals
	var numDigitsAfterDecimal int

	for i, s := range tickSize {
		if i > 1 {
			numDigitsAfterDecimal++
			if string(s) == "1" {
				break
			}
		}
	}

	// Round float to needed number of decimals
	priceString := data.RoundToPrecisionByString(priceFloat, numDigitsAfterDecimal)

	return priceString
}

// BinanceInPosition Checks if symbol has open position on Binance
func BinanceInPosition(client *futures.Client, symbol string) (isInPosition bool, position *futures.PositionRisk) {
	// Check binance this symbol has any p
	res, _ := client.NewGetPositionRiskService().Symbol(symbol).Do(context.Background())

	for _, p := range res {
		positionAmt, _ := strconv.ParseFloat(p.PositionAmt, 64)
		if positionAmt != 0 {
			return true, p
		}
	}

	return false, nil
}
