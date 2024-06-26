package main

import (
	"directionalMaker/data"
	"directionalMaker/oms"
	"directionalMaker/system"
	"log"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
)

func LiveTrader() {
	symbols := data.SymbolUSDTList([]string{
		"AGLD", "UNFI", "XLM", "1000PEPE", "SUI", "ARB", "APE",
	})

	var (
		apiKey    = "BinanceFuturesApiKey"
		secretKey = "BinanceFuturesSecretKey"
	)
	client := binance.NewFuturesClient(apiKey, secretKey)

	symbolsSettings := data.GetSymbolSettings(client, symbols)

	strategyList := system.GenerateStrategyList(symbolsSettings)

	// --------------------------------------------------------------
	// -----------------  LIVE TRADER
	chOrderset := make(chan *oms.OrderSet)
	chLastPrice := make(chan *data.LastPrice)
	chAggTrades := make(chan *futures.WsAggTradeEvent)

	// Event handler to send each new tick with aggTrade
	wsAggTradeHandler := func(event *futures.WsAggTradeEvent) {
		chAggTrades <- event
	}

	// Main Trading System Processor
	processor := system.NewProcessor(client, symbols, strategyList, chOrderset, chLastPrice)
	processor.Listen(chAggTrades)

	// OrderManagementSystem
	om := system.NewOMS(client, *strategyList, symbolsSettings, chOrderset, chLastPrice)
	om.Listen()

	// Start WS AggTrade service
	errHandler := func(err error) {
		log.Println(err)
	}

	doneC, _, err := futures.WsCombinedAggTradeServe(symbols, wsAggTradeHandler, errHandler)
	if err != nil {
		log.Println(err)
		return
	}

	<-doneC

}
