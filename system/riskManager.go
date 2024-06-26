package system

import (
	"context"
	"directionalMaker/data"
	"directionalMaker/oms"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"log"
	"strconv"
	"strings"
	"time"
)

// ResetInactiveSymbols Если на Binance нет позиции (Amount=0.0), но Symbol Busy
func (om *OMS) ResetInactiveSymbols() {

	om.inventory.Mu.Lock()
	om.openOrders.Mu.Lock()

	fmt.Println("\n===========================")
	fmt.Println("Checking Symbols Status")

	for symbol, symInv := range om.inventory.Stack {
		fmt.Println(data.SymbolNoUSDT(symbol), "busy:", om.openOrders.Busy[symbol], "|", "inv:", len(symInv.ActiveTrades))
	}

	for symbol, busy := range om.openOrders.Busy {

		// IF SYMBOL NOT BUSY & HAS ACTIVE TRADES
		if !busy && len(om.inventory.Stack[symbol].ActiveTrades) > 0 {
			// Check if it is in Binance position
			inPosition, _ := oms.BinanceInPosition(om.client, symbol)
			// If not in position - reset
			if !inPosition {
				log.Println("ResetInactiveSymbol:", symbol, "is Busy, but NOT in position.")
				// Delete from inventory
				om.inventory.Stack[symbol].ActiveTrades = []*oms.ActiveTrade{}
			}

		}

		// IF Symbol BUSY & NO ACTIVE TRADES
		if busy && len(om.inventory.Stack[symbol].ActiveTrades) == 0 {
			// Check if it is in Binance position
			inPosition, _ := oms.BinanceInPosition(om.client, symbol)
			// If not in position - reset
			if !inPosition {
				log.Println("ResetInactiveSymbol:", symbol, "is Busy, but NOT in position.")
				om.openOrders.SetBusy(symbol, false)
			}
		}

	}

	om.openOrders.Mu.Unlock()
	om.inventory.Mu.Unlock()

}

func ResetSymbol(symbol string) {
	// Reset Busy
	// Reset Active Trades
	// Cancel all orders on Binance
}

// -----------------------------
// O L D

// Looks up for symbols that are not busy, have no ActiveTrades, but in position. If such found - reset it.
func (om *OMS) rmCheckForAbandonedOrders() {
	// for list of symbols
	om.inventory.Mu.Lock()
	om.openOrders.Mu.Lock()

	for symbolKey, busy := range om.openOrders.Busy {

		// If symbol is not Busy, has zero ActiveTrades, but at the same time in position - that's bad.
		if !busy && len(om.inventory.Stack[symbolKey].ActiveTrades) == 0 {
			// Set symbol busy
			// Check binance this symbol has any p
			res, _ := om.client.NewGetPositionRiskService().Symbol(symbolKey).Do(context.Background())

			for _, p := range res {
				positionAmt, _ := strconv.ParseFloat(p.PositionAmt, 64)

				if positionAmt != 0 {
					// Cancel all open orders for this symbol
					om.client.NewCancelAllOpenOrdersService().Symbol(symbolKey).Do(context.Background())

					var marketExitSide oms.Side

					absQty := strings.ReplaceAll(p.PositionAmt, "-", "")

					if strings.Contains(p.PositionAmt, "-") {
						marketExitSide = oms.SideShort
					} else {
						marketExitSide = oms.SideLong
					}

					// MarketClose
					fmt.Println("MARKET CLOSE AbandonedPosition:")
					fmt.Println("Symbol:  ", symbolKey)
					fmt.Println("Quantity:", p.PositionAmt)

					_, err := om.client.NewCreateOrderService().Symbol(symbolKey).
						Side(oms.BinanceSide(marketExitSide)).Type(futures.OrderTypeMarket).Quantity(absQty).Do(context.Background())
					if err != nil {
						log.Println(err)
					}
				}
			}

			// Set symbol busy=false
			om.openOrders.SyncBusy(symbolKey, false)
		}
	}
	om.openOrders.Mu.Unlock()
	om.inventory.Mu.Unlock()
}

// If ActiveTrade has no Id for TakeProfit or StopLoss - it's a broken trade.
// We need to keep an eye on it, and close if it's in SL or TP price.
func (om *OMS) rmCheckForBrokenTradeSlTp(lastPrice *data.LastPrice) {
	hardValue := 1.005

	// For given Symbol, check each ActiveTrade in the inventory

	// Lock Mutex
	om.inventory.Mu.Lock()

	// Check if TP or SL has no ID
	for i, trade := range om.inventory.Stack[lastPrice.Symbol].ActiveTrades {
		var idToCancel int64
		var marketOrderSide oms.Side

		// if NO SL: check if lastPrice is 0.5% less than EntryPrice: if yes - MarketClose
		if trade.BinanceIdSL < 10 && time.Now().UnixMilli() > (trade.EntryTime+7*time.Second.Milliseconds()) {
			// For Long trade, if price <= hardStop: exit
			if trade.Side == oms.SideLong {
				hardStop := trade.EntryPrice / hardValue
				if lastPrice.Close <= hardStop {
					// Exit
					marketOrderSide = oms.SideShort

					// Id of TP
					idToCancel = trade.BinanceIdTP
				}
				// For Short trade, if price >= hardStop: exit
			} else if trade.Side == oms.SideShort {
				hardStop := trade.EntryPrice * hardValue
				if lastPrice.Close >= hardStop {
					// Exit
					marketOrderSide = oms.SideLong

					// Id of TP
					idToCancel = trade.BinanceIdTP
				}
			}
		}

		// if NO TP: check if lastPrice is 0.5% more than EntryPrice: if yes - MarketClose
		if trade.BinanceIdTP < 10 && time.Now().UnixMilli() > (trade.EntryTime+7*time.Second.Milliseconds()) {
			// For Long trade, if price >= hardTakeProfit: exit
			if trade.Side == oms.SideLong {
				hardTakeProfit := trade.EntryPrice * hardValue
				if lastPrice.Close >= hardTakeProfit {
					// Exit
					marketOrderSide = oms.SideShort

					// Id of SL
					idToCancel = trade.BinanceIdSL

				}
				// For Short trade, if price <= hardTakeProfit: exit
			} else if trade.Side == oms.SideShort {
				hardTakeProfit := trade.EntryPrice / hardValue
				if lastPrice.Close <= hardTakeProfit {
					// Exit
					marketOrderSide = oms.SideLong

					// Id of SL
					idToCancel = trade.BinanceIdSL

				}
			}
		}

		// If broken order was found: market exit one order, and cancel another.
		if (trade.BinanceIdSL < 10 || trade.BinanceIdTP < 10) && time.Now().UnixMilli() > (trade.EntryTime+7*time.Second.Milliseconds()) {
			fmt.Println("(!) RISK MANAGER EXITS", lastPrice.Symbol, "POSITION")

			var qty string
			for _, s := range om.symbolsSettings {
				if s.Symbol == lastPrice.Symbol {
					qty = fmt.Sprintf("%f", s.Quantity)
				}
			}

			// Market Close selected order
			// TODO - confirmation of success
			if qty != "" {
				_, err := om.client.NewCreateOrderService().Symbol(lastPrice.Symbol).
					Side(oms.BinanceSide(marketOrderSide)).Type(futures.OrderTypeMarket).
					TimeInForce(futures.TimeInForceTypeGTC).Quantity(qty).Do(context.Background())
				if err != nil {
					log.Println(err)
				}
			}

			// Cancel the other one (tp or sl)
			//_, err := om.client.NewCancelOrderService().Symbol(lastPrice.Symbol).OrderID(idToCancel).Do(context.Background())
			//if err != nil {
			//	log.Println("CancelMultipleOrders:", err)
			//}
			// TODO : Временно - отменяем все заказы на символе
			err := om.client.NewCancelAllOpenOrdersService().Symbol(lastPrice.Symbol).Do(context.Background())
			if err != nil {
				log.Println("CancelMultipleOrders:", err)
				fmt.Println(idToCancel)
			}

			// Delete from inventory
			om.inventory.Stack[lastPrice.Symbol].RemoveActiveOrder(i)

			// Set symbol Busy=false
			om.openOrders.SyncBusy(lastPrice.Symbol, false)
		}

	}
	// Unlock Mutex
	om.inventory.Mu.Unlock()
}
