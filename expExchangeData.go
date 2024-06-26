package main

import (
	"directionalMaker/data"
	"directionalMaker/oms"
	"fmt"
	"github.com/adshao/go-binance/v2"
)

func expExchangeData() {
	var (
		apiKey    = "CZYRAfunS6FVDQJH5cjZCQsA89SsiJrVqCN5G5DR8HGzyiVJXPw2k4ASChjR6Lgk"
		secretKey = "lGvXuV3iaYDo4LovFu03nOPCKguCFo83M2uZpVxRe9ANGcsPsyo0akZJOP7mVsKS"
	)
	client := binance.NewFuturesClient(apiKey, secretKey)
	//
	//res, err := client.NewExchangeInfoService().Do(context.Background())
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//for _, s := range res.Symbols {
	//	if s.Symbol == "APT"+"USDT" {
	//		fmt.Println("SYMBOL:", s.Symbol)
	//		fmt.Println("PricePrecision:    ", s.PricePrecision)
	//		fmt.Println("QuotePrecision:    ", s.QuotePrecision)
	//		fmt.Println("QuantityPrecision: ", s.QuantityPrecision)
	//		fmt.Println("BaseAssetPrecision:", s.BaseAssetPrecision)
	//		fmt.Println("PriceFilter:       ", s.PriceFilter().TickSize)
	//	}
	//}

	symbols := data.SymbolUSDTList([]string{
		"APT", "GMT", "OP", "APE", "WAVES", "GALA", "SOL", "1000SHIB", "CHZ", "NEAR",
	})

	//data.GetSymbolSettings(client, symbols)
	ss := data.GetSymbolSettings(client, symbols)
	for _, s := range ss {
		fmt.Println()
		fmt.Println(s.Symbol)
		fmt.Println("Quantity", s.Quantity)
		fmt.Println("QuantityPrecision", s.QuantityPrecision)
		fmt.Println("TickSize", s.TickSize)
	}

}

func expTestingConvertFloatToBinancePrice() {
	//priceDesired := "2.3997"
	priceFloat := 2.39970103123
	tickSize := "0.0001000"

	priceString := oms.BinancePrice(priceFloat, tickSize)
	fmt.Println(priceString)

}
