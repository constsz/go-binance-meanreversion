package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
)

func expGetSymbolOrders() {
	var (
		apiKey    = "CZYRAfunS6FVDQJH5cjZCQsA89SsiJrVqCN5G5DR8HGzyiVJXPw2k4ASChjR6Lgk"
		secretKey = "lGvXuV3iaYDo4LovFu03nOPCKguCFo83M2uZpVxRe9ANGcsPsyo0akZJOP7mVsKS"
	)
	client := binance.NewFuturesClient(apiKey, secretKey)

	//res, _ := client.NewListOpenOrdersService().Symbol("OPUSDT").Do(context.Background())
	//
	//for i, v := range res {
	//	fmt.Println(i)
	//	fmt.Println(v)
	//}

	res, _ := client.NewGetPositionRiskService().Symbol("APTUSDT").Do(context.Background())

	for i, position := range res {
		fmt.Println(i)
		fmt.Println("Symbol:          ", position.Symbol)
		fmt.Println("EntryPrice:      ", position.EntryPrice)
		fmt.Println("PositionAmt:     ", position.PositionAmt) // no position = 0.0
		fmt.Println("PositionSide:    ", position.PositionSide)
		fmt.Println("UnRealizedProfit:", position.UnRealizedProfit)
	}

	//res2, _ := client.NewListOrdersService().Symbol("OPUSDT").Do(context.Background())
	//
	//for i, v := range res2 {
	//	if v.Status == futures.OrderStatusTypeFilled {
	//		fmt.Println(i)
	//		fmt.Println(v)
	//	}
	//}
}
