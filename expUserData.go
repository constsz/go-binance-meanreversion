package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"log"
)

func expUserData() {
	var (
		apiKey    = "q8HGCy5PRwh3NiwWA1iU2XCpRacCysravGkWPi9lLdFJzfBEpYakJwHDTofWAUA2"
		secretKey = "HswzLFhuc9poxfD5uXJdkWSZ8ke838PJeJOcylIHWjj4pE7X2ScVwVuLvIYg8ECr"
	)

	client := binance.NewFuturesClient(apiKey, secretKey)

	// Get ListenKey from Binance to use for WsUserData
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	// WS User Data Listener
	errHandler := func(err error) {
		log.Println("WsUserDataService: ", err)
	}

	wsUserHandler := func(ude *futures.WsUserDataEvent) {
		fmt.Println()
		log.Println("processUserDataEvent:")
		fmt.Println("Symbol:         ", ude.OrderTradeUpdate.Symbol) // GMTUSDT
		//fmt.Println("TradeID:      ", ude.OrderTradeUpdate.TradeID)       // 584246786 (if filled)
		//fmt.Println("ClientOrderID:", ude.OrderTradeUpdate.ClientOrderID) // x-Jbzjn7gQ0KuB1NVxl0isimHVh1Xj8A
		fmt.Println("ActivationPrice:", ude.OrderTradeUpdate.ActivationPrice)      //
		fmt.Println("OriginalPrice:  ", ude.OrderTradeUpdate.OriginalPrice)        //
		fmt.Println("AveragePrice:   ", ude.OrderTradeUpdate.AveragePrice)         //
		fmt.Println("LastFilledPrice:", ude.OrderTradeUpdate.LastFilledPrice)      //
		fmt.Println("ID:           	 ", ude.OrderTradeUpdate.ID)                   // 10708668898
		fmt.Println("Side:         	 ", ude.OrderTradeUpdate.Side)                 // BUY, SELL
		fmt.Println("Status:         ", ude.OrderTradeUpdate.Status)               // NEW, CANCELLED, FILLED
		fmt.Println("Type:           ", ude.OrderTradeUpdate.Type)                 // LIMIT
		fmt.Println("Event:          ", ude.Event)                                 // LIMIT
		fmt.Println("LastFilledQty:  ", ude.OrderTradeUpdate.LastFilledQty)        // LIMIT
		fmt.Println("LastFilledQty:  ", ude.OrderTradeUpdate.AccumulatedFilledQty) // LIMIT
		fmt.Println()
	}

	doneC, _, err := futures.WsUserDataServe(listenKey, wsUserHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	<-doneC

}

func expWsTrades() {
	wsAggTradeHandler := func(event *futures.WsAggTradeEvent) {
		fmt.Println(event)
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, _, err := futures.WsCombinedAggTradeServe([]string{"OPUSDT", "SOLUSDT"}, wsAggTradeHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	<-doneC
}

// Create Multiple Orders
// LIMIT = 5 orders
func expCreateOrders() {
	// MATICUSDT  long, qty=4 	: 1.3780, 1.3750, 1.3720,
	// GALAUSDT   long, qty=120 : 0.04460, 0.04440, 0.04420,
	var (
		apiKey    = "q8HGCy5PRwh3NiwWA1iU2XCpRacCysravGkWPi9lLdFJzfBEpYakJwHDTofWAUA2"
		secretKey = "HswzLFhuc9poxfD5uXJdkWSZ8ke838PJeJOcylIHWjj4pE7X2ScVwVuLvIYg8ECr"
	)

	client := binance.NewFuturesClient(apiKey, secretKey)

	order1 := client.NewCreateOrderService().Symbol("APTUSDT").
		Side(futures.SideTypeBuy).Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).Quantity("0.8").
		Price("11.590")
	//order2 := client.NewCreateOrderService().Symbol("MATICUSDT").
	//	Side(futures.SideTypeBuy).Type(futures.OrderTypeLimit).
	//	TimeInForce(futures.TimeInForceTypeGTC).Quantity("4").
	//	Price("1.3750")
	//order3 := client.NewCreateOrderService().Symbol("MATICUSDT").
	//	Side(futures.SideTypeBuy).Type(futures.OrderTypeLimit).
	//	TimeInForce(futures.TimeInForceTypeGTC).Quantity("4").
	//	Price("1.3720")
	//
	//order4 := client.NewCreateOrderService().Symbol("GALAUSDT").
	//	Side(futures.SideTypeBuy).Type(futures.OrderTypeLimit).
	//	TimeInForce(futures.TimeInForceTypeGTC).Quantity("120").
	//	Price("0.04460")
	//order5 := client.NewCreateOrderService().Symbol("GALAUSDT").
	//	Side(futures.SideTypeBuy).Type(futures.OrderTypeLimit).
	//	TimeInForce(futures.TimeInForceTypeGTC).Quantity("120").
	//	Price("0.04440")

	orders := []*futures.CreateOrderService{order1 /*, order2, order3, order4, order5*/}

	batchOrders, err := client.NewCreateBatchOrdersService().OrderList(orders).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, order := range batchOrders.Orders {
		fmt.Println(order.OrderID)
		fmt.Println("OrigQuantity", order.OrigQuantity)
		fmt.Println("CumQuantity", order.CumQuantity)
		fmt.Println("CumQuote", order.CumQuote)
		fmt.Println("ExecutedQuantity", order.ExecutedQuantity)
	}

}

// Cancel multiple Orders
func expCancelListOfOrders() {
	// MATICUSDT  long, qty=4 : 1.3780, 1.3750, 1.3720,
	var (
		apiKey    = "q8HGCy5PRwh3NiwWA1iU2XCpRacCysravGkWPi9lLdFJzfBEpYakJwHDTofWAUA2"
		secretKey = "HswzLFhuc9poxfD5uXJdkWSZ8ke838PJeJOcylIHWjj4pE7X2ScVwVuLvIYg8ECr"
	)

	client := binance.NewFuturesClient(apiKey, secretKey)

	//cancelsListMATIC := []int64{24086974992, 24086974991, 24086974993}
	cancelsListGALA := []int64{9077458793, 9077458794}

	cancelledMultiple, err := client.NewCancelMultipleOrdersService().
		Symbol("GALAUSDT").OrderIDList(cancelsListGALA).
		//Symbol("MATICUSDT").OrderIDList(cancelsListMATIC).
		Do(context.Background())

	//cancelled, err := client.NewCancelOrderService().Symbol("GALAUSDT").OrderID(13123123).Do()

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, order := range cancelledMultiple {
		fmt.Println(order)
	}

}
