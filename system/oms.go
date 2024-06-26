package system

import (
	"context"
	"directionalMaker/data"
	"directionalMaker/oms"
	"directionalMaker/ta"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"log"
	"strconv"
	"sync"
	"time"
)

// TODO : DEBUG
var OrderBatchPostingTimeFrequency = 200

//var AbandonedOrdersCheckingFrequency = 7

type OMS struct {
	chOrderset  <-chan *oms.OrderSet
	chLastPrice <-chan *data.LastPrice
	chUserData  chan *futures.WsUserDataEvent

	symbolInfo  SymbolInfo
	symbolsList []string
	inventory   *oms.Inventory
	openOrders  *oms.OpenOrders

	symbolsSettings data.SymbolsSettings
	client          *futures.Client
}

type SymbolInfo struct {
	Mu         sync.Mutex
	LastPrices map[string]float64
}

func (sinf *SymbolInfo) UpdateLastPrice(symbol string, price float64) {
	sinf.Mu.Lock()
	sinf.LastPrices[symbol] = price
	sinf.Mu.Unlock()
}

func NewOMS(client *futures.Client, strategyList StrategyList, symbolsSettings data.SymbolsSettings, chOrderset <-chan *oms.OrderSet, chLastPrice <-chan *data.LastPrice) *OMS {
	lastPrices := make(map[string]float64, len(strategyList))

	var symbolsList []string
	for symbol := range strategyList {
		symbolsList = append(symbolsList, symbol)
	}

	newOMS := &OMS{
		chOrderset:  chOrderset,
		chLastPrice: chLastPrice,
		chUserData:  make(chan *futures.WsUserDataEvent),
		inventory:   oms.NewInventory(&symbolsSettings, 0),
		openOrders:  oms.NewOpenOrders(symbolsList),
		symbolInfo:  SymbolInfo{LastPrices: lastPrices},
		symbolsList: symbolsList,

		symbolsSettings: symbolsSettings,
		client:          client,
	}

	return newOMS
}

func (om *OMS) Listen() {
	tickerOrderBatchPosting := time.NewTicker(time.Duration(OrderBatchPostingTimeFrequency) * time.Millisecond)
	tickerResetInactiveSymbols := time.NewTicker(2 * time.Minute)

	go userDataListener(om.client, om.chUserData)

	timeStarted := time.Now().UnixMilli()
	var timePassed bool

	go func() {
		for {
			if !timePassed && time.Now().UnixMilli() > timeStarted+15*time.Second.Milliseconds() {
				timePassed = true
			}

			select {
			// UserDataEvents: OrderFilled (openOrder entry, TP, SL)
			// В самом начале, тк наивысший приоритет
			case ude := <-om.chUserData:
				om.processUserDataEvent(ude)
			// New OrderSet from strategy
			case orderset := <-om.chOrderset:
				om.openOrders.AddNew(orderset.Symbol, orderset)
			// Batch order poster by timer
			case <-tickerOrderBatchPosting.C:
				if timePassed {
					om.batchOrderPoster()
				}
			case <-tickerResetInactiveSymbols.C:
				om.ResetInactiveSymbols()
			}
		}
	}()
}

// Each 200ms or so, grab not-yet-posted OpenOrders from the stack
// if the symbol is not Busy already.
// and batch update them on Binance (delete & create),
// and then replace old StackPosted with the updated one.
func (om *OMS) batchOrderPoster() {
	om.openOrders.Mu.Lock()
	om.inventory.Mu.Lock()

	// Grab new openOrders
	newOpenOrders := om.GrabNewOpenOrders()

	// If new Open Orders available:
	if len(*newOpenOrders) > 0 {
		//fmt.Println("\nnew open orders symbols:")
		//fmt.Println(newOpenOrders.ToSymbolList(), "\n")

		// Grab postedOrders for same symbols as available new openOrders
		// Sets symbol to Busy=true
		postedOpenOrders := om.openOrders.GrabPosted(newOpenOrders)

		// Batch update orders on Binance (delete and create again)
		newPostedOpenOrders := om.binanceBatchUpdateOrders(newOpenOrders, postedOpenOrders)

		// Write new posted open orders into the StackPosted.
		if len(*newPostedOpenOrders) > 0 {
			// Add newly posted orders to StackPosted and Set symbols to Busy=false
			om.openOrders.UpdateStackPosted(newPostedOpenOrders)

		} else {
			om.openOrders.SetMultipleBusy(newOpenOrders, false)
			log.Println("batchOrderPoster: Got empty slice from Binance |", newOpenOrders.ToSymbolList())
		}

	}

	om.openOrders.Mu.Unlock()
	om.inventory.Mu.Unlock()

}

// GrabNewOpenOrders grabs all new non-blank openOrders from Stack and resets the StackNew
// Sets Busy to true, because if it grabbed the order, it will release Busy only
// when final step ends.
func (om *OMS) GrabNewOpenOrders() *oms.OpenOrderStack {
	grabbedOpenOrders := make(oms.OpenOrderStack)

	for symbolKey, openOrder := range om.openOrders.StackNew {
		// Check if Capacity allows for new trades for the Symbol
		if len(om.inventory.Stack[symbolKey].ActiveTrades) < om.inventory.Capacity {
			// Check if element of StackNew is not reset before
			// Check if Symbol is not already Busy
			if openOrder.State != oms.OpenOrderBlank && !om.openOrders.Busy[symbolKey] {
				// TODO: Print before and after to check if they are really reset to nil
				grabbedOpenOrders[symbolKey] = openOrder

				om.openOrders.Busy[symbolKey] = true
				om.openOrders.StackNew[symbolKey] = &oms.OpenOrder{}
			}
		}
	}

	return &grabbedOpenOrders
}

func (om *OMS) binanceBatchUpdateOrders(newOpenOrders, postedOpenOrders *oms.OpenOrderStack) *oms.OpenOrderStack {
	// First we cancel old OpenOrders from StackPosted
	if len(*postedOpenOrders) > 0 {
		om.CancelMultipleOrders(postedOpenOrders)
	}

	// Now create new orders
	newPostedOrders := om.CreateMultipleOrders(newOpenOrders)

	return newPostedOrders

}

func (om *OMS) CreateMultipleOrders(newOpenOrders *oms.OpenOrderStack) *oms.OpenOrderStack {
	step := 5

	// Fill this slice with batchesOfOrders of orders: [[1,2,3,4,5], [6,7,8,9,10], ...]
	var batchesOfOrders [][]*futures.CreateOrderService

	i := 1
	var singleBatch []*futures.CreateOrderService
	for symbol, order := range *newOpenOrders {
		//fmt.Println("CreateMultipleOrders length of inventory:\n", len(om.inventory.Stack[symbol].ActiveTrades))
		side := oms.BinanceSide(order.Side)
		if side == "" {
			log.Println("om.CreateMultipleOrders: error during call oms.BinanceSide -- wrong returned value.")
		}

		price := oms.BinancePrice(order.EntryPrice, om.symbolsSettings.SymbolTickSize(symbol))
		qty := fmt.Sprintf("%f", order.Quantity)

		newOrder := om.client.NewCreateOrderService().Symbol(symbol).
			Side(side).Type(futures.OrderTypeLimit).
			TimeInForce(futures.TimeInForceTypeGTC).Quantity(qty).
			Price(price)

		singleBatch = append(singleBatch, newOrder)

		if i == step || i == len(*newOpenOrders) {
			batchesOfOrders = append(batchesOfOrders, singleBatch)
			singleBatch = []*futures.CreateOrderService{}
		}

		i++

	}

	newPostedOrders := make(oms.OpenOrderStack)

	type newPostedOrdersSyncType struct {
		mu    sync.Mutex
		items oms.OpenOrderStack
	}

	newPostedOrdersSync := newPostedOrdersSyncType{
		items: newPostedOrders,
	}

	// Send each batch of orders to Binance
	var wg sync.WaitGroup
	wg.Add(len(batchesOfOrders))
	for j, batch := range batchesOfOrders {
		if j > 0 {
			time.Sleep(10 * time.Millisecond)
		}
		go func(batch []*futures.CreateOrderService, newPostedOrdersSync *newPostedOrdersSyncType) {
			batchOrders, err := om.client.NewCreateBatchOrdersService().OrderList(batch).Do(context.Background())
			if err != nil {
				fmt.Println("CreateMultipleOrders:", err)
			} else {
				// Aggregate new StackPosted with these new posted orders
				for _, order := range batchOrders.Orders {
					p, _ := strconv.ParseFloat(order.Price, 10)
					q, _ := strconv.ParseFloat(order.OrigQuantity, 10)

					newPostedOrdersSync.mu.Lock()
					newPostedOrdersSync.items[order.Symbol] = &oms.OpenOrder{
						State:      oms.OpenOrderPosted,
						BinanceId:  order.OrderID,
						Side:       oms.BinanceSideToInternal(order.Side),
						EntryPrice: p,
						Quantity:   q,
						TP:         (*newOpenOrders)[order.Symbol].TP,
						SL:         (*newOpenOrders)[order.Symbol].SL,
					}
					newPostedOrdersSync.mu.Unlock()
				}
			}
			wg.Done()
		}(batch, &newPostedOrdersSync)
	}
	wg.Wait()

	finalPostedOrders := newPostedOrdersSync.items

	return &finalPostedOrders
}

type processedOrders struct {
	Mu    sync.Mutex
	Items *[]oms.OpenOrder
}

// CreateTpSlOrders creates orders on Binance for executed openOrder.
// It returns array[2] { 0: TP order, 1: SL order}
func (om *OMS) CreateTpSlOrders(tpOrder, slOrder *oms.OpenOrder, symbol string) *processedOrders {

	tpPrice := oms.BinancePrice(tpOrder.EntryPrice, om.symbolsSettings.SymbolTickSize(symbol))
	tpQty := fmt.Sprintf("%f", tpOrder.Quantity)

	slPrice := oms.BinancePrice(slOrder.EntryPrice, om.symbolsSettings.SymbolTickSize(symbol))
	slQty := fmt.Sprintf("%f", slOrder.Quantity)

	tpOrderToPost := om.client.NewCreateOrderService().Symbol(symbol).
		Side(oms.BinanceSide(tpOrder.Side)).Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).Quantity(tpQty).Price(tpPrice)
	slOrderToPost := om.client.NewCreateOrderService().Symbol(symbol).
		Side(oms.BinanceSide(slOrder.Side)).Type(futures.OrderTypeStopMarket).
		TimeInForce(futures.TimeInForceTypeGTC).Quantity(slQty). /*.Price(slPrice)*/ StopPrice(slPrice)
	var orderNil *futures.CreateOrderResponse

	orders := []*futures.CreateOrderService{tpOrderToPost, slOrderToPost}

	processedItemsSlice := make([]oms.OpenOrder, 2)
	processedOrdersTpSl := processedOrders{
		Items: &processedItemsSlice,
	}

	var wg sync.WaitGroup
	// Run loop to process orders concurrently
	wg.Add(len(orders))
	for i, order := range orders {
		if i > 0 {
			time.Sleep(10 * time.Millisecond)
		}
		go func(i int, order *futures.CreateOrderService, processedOrdersTpSl *processedOrders) {

			var tryes int
		tryAgain:
			res, err := order.Do(context.Background())
			if err != nil {
				log.Println("CreateTpSlOrders:", err)
			}

			// Check if in response we got empty order
			if res == orderNil {
				log.Println("CreateTpSlOrders : EMPTY RES! Trying again.")
				tryes++
				if tryes < 3 {
					time.Sleep(1 * time.Second)
					goto tryAgain
				} else {
					// If after 3 tries we still have no
				}
			} else {
				p, _ := strconv.ParseFloat(res.Price, 10)
				q, _ := strconv.ParseFloat(res.OrigQuantity, 10)

				openOrder := oms.OpenOrder{
					State:      oms.OpenOrderPosted,
					BinanceId:  res.OrderID,
					Side:       oms.BinanceSideToInternal(res.Side),
					EntryPrice: p,
					Quantity:   q,
				}
				processedOrdersTpSl.Mu.Lock()
				(*processedOrdersTpSl.Items)[i] = openOrder
				processedOrdersTpSl.Mu.Unlock()
			}
			wg.Done()

		}(i, order, &processedOrdersTpSl)
	}
	wg.Wait()
	return &processedOrdersTpSl

}

func (om *OMS) CancelMultipleOrders(postedOpenOrders *oms.OpenOrderStack) {
	if len(*postedOpenOrders) > 0 {
		var wg sync.WaitGroup
		var i int
		wg.Add(len(*postedOpenOrders))
		for symbol, order := range *postedOpenOrders {
			if i > 0 {
				time.Sleep(3 * time.Millisecond)
			}
			i++
			go func(symbol string, order *oms.OpenOrder, wg *sync.WaitGroup) {
				err := om.client.NewCancelAllOpenOrdersService().Symbol(symbol).Do(context.Background())
				if err != nil {
					log.Println("CancelMultipleOrders:", err)
				}
				wg.Done()
			}(symbol, order, &wg)

		}
		wg.Wait()
	}
}

// processUserDataEvent | Listens to Order Fills
// If order was Filled:
//   - from OpenOrders = new trade
//   - from Inventory  = closing position, TP or SL
//
// First we check case for OpenOrder, because it needs fast TP and SL orders.
// If it is the case - function returns. Next we do same checking for Inventory.ActiveTrades,
// because this is less important (as it's just a position exit).
//
// :: futures/websocket_service.go :: WsOrderTradeUpdate struct
func (om *OMS) processUserDataEvent(ude *futures.WsUserDataEvent) {
	// Check if Event type is OrderTradeUpdate
	if ude.Event == futures.UserDataEventTypeOrderTradeUpdate {
		upd := ude.OrderTradeUpdate

		// Check if update is for one of symbols from our list
		var correctSymbol bool
		for _, symbol := range om.symbolsList {
			if upd.Symbol == symbol {
				correctSymbol = true
			}
		}

		if correctSymbol {

			// If update.Status == FILLED
			if upd.Status == futures.OrderStatusTypeFilled {
				func() {
					fmt.Println()
					log.Println("processUserDataEvent:")
					fmt.Println("Symbol:       ", ude.OrderTradeUpdate.Symbol) // GMTUSDT
					//fmt.Println("TradeID:      ", ude.OrderTradeUpdate.TradeID)       // 584246786 (if filled)
					//fmt.Println("ClientOrderID:", ude.OrderTradeUpdate.ClientOrderID) // x-Jbzjn7gQ0KuB1NVxl0isimHVh1Xj8A
					fmt.Println("ID:           ", ude.OrderTradeUpdate.ID)            // 10708668898
					fmt.Println("Side:         ", ude.OrderTradeUpdate.Side)          // BUY, SELL
					fmt.Println("Status:       ", ude.OrderTradeUpdate.Status)        // NEW, CANCELLED, FILLED
					fmt.Println("Type:         ", ude.OrderTradeUpdate.Type)          // LIMIT
					fmt.Println("Event:        ", ude.Event)                          // LIMIT
					fmt.Println("LastFilledQty:", ude.OrderTradeUpdate.LastFilledQty) // LIMIT
					fmt.Println()
				}()

				om.inventory.Mu.Lock()
				om.openOrders.Mu.Lock()
				om.openOrders.SetBusy(upd.Symbol, true)

				// Cancel before adding TP SL
				err := om.client.NewCancelAllOpenOrdersService().Symbol(upd.Symbol).Do(context.Background())
				if err != nil {
					log.Println("CreateTpSlOrders: CANCEL Orders:", err)
				}

				// if len(ActiveTrades) > 0 -> we EXIT position, if it's == 0: we enter NEW position
				// >>> EXIT
				if len(om.inventory.Stack[upd.Symbol].ActiveTrades) > 0 {

					// Часто бывает, видимо пока Binance не успел обновить размещенные заказы,
					// может открыться сразу несколько заказов, видимо пара старых и новый.
					//
					// Первый пришел, его добавили в Inventory ActiveTrades.
					// А второй когда пришел, его записали в Exit, тк бот подумал, что раз ордер filled, значит это exit.
					//
					// Если ActiveTrades есть, и приходит новый Filled,
					// нужно посмотреть Side:
					//		- Если Side противоположный, то это Exit.
					//		- Если Side == ActiveTrade.Side - то нужно увеличить Quantity, отменить все
					//		  ордера и выставить новые TP/SL с новыми Quantity.

					//if oms.BinanceSideToInternal(upd.Side) == activeTrade.Side {}

					// If ActiveTrade was filled => Cancel TP or SL
					// Check Inventory[symbol].ActiveTrades : IdSL and IdTP
					// If it is - delete the order from Inventory.Stack[symbol].ActiveTrades

					symbolInventory := om.inventory.Stack[upd.Symbol]

					// Find if activeTrade exists
					var idToDelete int64
					var indexToDelete int
					var activeTradeInPosition *oms.ActiveTrade

					for i, actTr := range symbolInventory.ActiveTrades {
						activeTradeInPosition = actTr
						if upd.ID == actTr.BinanceIdSL {
							idToDelete = actTr.BinanceIdTP
							indexToDelete = i
						} else if upd.ID == actTr.BinanceIdTP {
							idToDelete = actTr.BinanceIdSL
							indexToDelete = i
						}
					}

					// if activeTrade exists -> check fill Side
					// if side is same as ActiveTrade -> add quantity, cancel all orders and add new TpSl
					// if side is opposite -> this is Exit
					if activeTradeInPosition.Side == oms.BinanceSideToInternal(upd.Side) {
						fmt.Println("\n>>>>> ADD TO POSITION\n")

						fmt.Println("\n~~~~~~~~~~~~~~~~~~~~~~")
						fmt.Println("ActiveTrade Check (BEFORE):")
						fmt.Println("Side:       ", activeTradeInPosition.Side)
						fmt.Println("Quantity:   ", activeTradeInPosition.Quantity)
						fmt.Println("BinanceIdTP:", activeTradeInPosition.BinanceIdTP)
						fmt.Println("BinanceIdSL:", activeTradeInPosition.BinanceIdSL)

						// Cancel all orders
						err = om.client.NewCancelAllOpenOrdersService().Symbol(upd.Symbol).Do(context.Background())

						q, _ := strconv.ParseFloat(upd.LastFilledQty, 64)

						// Add quantity to ActiveTrade
						activeTradeInPosition.Quantity += q

						tpSide := oms.SideShort
						slSide := oms.SideShort
						if activeTradeInPosition.Side == oms.SideShort {
							tpSide = oms.SideLong
							slSide = oms.SideLong
						}

						// Calculate new TakeProfit and StopLoss Prices
						p, _ := strconv.ParseFloat(upd.LastFilledPrice, 64)

						// LONG TP,SL
						tp := p * ta.Pct(om.openOrders.StackPosted[upd.Symbol].TP)
						sl := p / ta.Pct(om.openOrders.StackPosted[upd.Symbol].SL)

						// SHORT TP,SL
						if activeTradeInPosition.Side == oms.SideShort {
							tp = p / ta.Pct(om.openOrders.StackPosted[upd.Symbol].TP)
							sl = p * ta.Pct(om.openOrders.StackPosted[upd.Symbol].SL)
						}

						tpOrder := &oms.OpenOrder{
							State:      oms.OpenOrderNew,
							Side:       tpSide,
							EntryPrice: tp,
							Quantity:   activeTradeInPosition.Quantity,
						}

						slOrder := &oms.OpenOrder{
							State:      oms.OpenOrderNew,
							Side:       slSide,
							EntryPrice: sl,
							Quantity:   activeTradeInPosition.Quantity,
						}

						postedTpSlOrders := om.CreateTpSlOrders(tpOrder, slOrder, upd.Symbol)

						postedTP := (*postedTpSlOrders.Items)[0]
						postedSL := (*postedTpSlOrders.Items)[1]

						// 		- create ActiveTrade; fill Binance Ids for Entry, TP, SL, TradeId; append to Inventory.Stack[symbol]
						activeTradeInPosition.BinanceIdTP = postedTP.BinanceId
						activeTradeInPosition.BinanceIdSL = postedSL.BinanceId

						fmt.Println("\nActiveTrade Check (AFTER):")
						fmt.Println("Side:       ", activeTradeInPosition.Side)
						fmt.Println("Quantity:   ", activeTradeInPosition.Quantity)
						fmt.Println("BinanceIdTP:", activeTradeInPosition.BinanceIdTP)
						fmt.Println("BinanceIdSL:", activeTradeInPosition.BinanceIdSL)
						fmt.Println()

					} else if activeTradeInPosition.Side != oms.BinanceSideToInternal(upd.Side) {
						fmt.Println("\n>>>>> EXIT\n")

						if idToDelete > 0 {

							err = om.client.NewCancelAllOpenOrdersService().Symbol(upd.Symbol).Do(context.Background())
							if err != nil {
								log.Println("processUserDataEvent: Exit order: NewCancelOrderService: ", err)
							} else {
								symbolInventory.RemoveActiveOrder(indexToDelete)
							}
						}

					}

				} else {
					// >>> ENTRY
					// If this fill is not Exiting existing trade, it means we open new position
					// - send TP and SL orders
					positionSide := oms.BinanceSideToInternal(upd.Side)
					// Set Side of TP and SL (opposite to Entry)
					tpSide := oms.SideShort
					slSide := oms.SideShort
					if positionSide == oms.SideShort {
						tpSide = oms.SideLong
						slSide = oms.SideLong
					}

					// Calculate Stoploss and TakeProfit
					p, _ := strconv.ParseFloat(upd.LastFilledPrice, 64)
					q := om.openOrders.StackPosted[upd.Symbol].Quantity

					// LONG TP,SL
					tp := p * ta.Pct(om.openOrders.StackPosted[upd.Symbol].TP)
					sl := p / ta.Pct(om.openOrders.StackPosted[upd.Symbol].SL)

					// SHORT TP,SL
					if positionSide == oms.SideShort {
						tp = p / ta.Pct(om.openOrders.StackPosted[upd.Symbol].TP)
						sl = p * ta.Pct(om.openOrders.StackPosted[upd.Symbol].SL)
					}

					tpOrder := &oms.OpenOrder{
						State:      oms.OpenOrderNew,
						Side:       tpSide,
						EntryPrice: tp,
						Quantity:   q,
					}

					slOrder := &oms.OpenOrder{
						State:      oms.OpenOrderNew,
						Side:       slSide,
						EntryPrice: sl,
						Quantity:   q,
					}

					postedTpSlOrders := om.CreateTpSlOrders(tpOrder, slOrder, upd.Symbol)

					postedTP := (*postedTpSlOrders.Items)[0]
					postedSL := (*postedTpSlOrders.Items)[1]

					// 		- create ActiveTrade; fill Binance Ids for Entry, TP, SL, TradeId; append to Inventory.Stack[symbol]
					newActiveTrade := &oms.ActiveTrade{
						Side:        positionSide,
						EntryPrice:  p,
						EntryTime:   upd.TradeTime,
						Quantity:    q,
						TP:          postedTP.EntryPrice,
						SL:          postedSL.EntryPrice,
						BinanceIdTP: postedTP.BinanceId,
						BinanceIdSL: postedSL.BinanceId,
					}

					// Add new ActiveTrade to Inventory
					om.inventory.InsertActiveTrade(upd.Symbol, newActiveTrade)

				}

				// Release the Symbol Busy to false
				om.openOrders.SetBusy(upd.Symbol, false)
				om.inventory.Mu.Unlock()
				om.openOrders.Mu.Unlock()

			}
		}

	}

	// PS: also add logic to deal with futures.OrderStatusTypePartiallyFilled

}

func userDataListener(client *futures.Client, chUserData chan<- *futures.WsUserDataEvent) {
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

	wsUserHandler := func(event *futures.WsUserDataEvent) {
		chUserData <- event
	}

	doneC, _, err := futures.WsUserDataServe(listenKey, wsUserHandler, errHandler)

	tickerKeepAlive := time.NewTicker(3 * time.Minute)

	go func() {
		for {
			<-tickerKeepAlive.C
			fmt.Println("<><><><> KeepAlive UserStreamService")
			client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(context.Background())
		}
	}()

	if err != nil {
		fmt.Println(err)
		return
	}
	<-doneC
}
