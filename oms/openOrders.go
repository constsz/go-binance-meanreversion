package oms

import (
	"sync"
)

// OpenOrders used as a buffer
// - strategy sends new ordersets that overwrite existing ones
// in StackNew
// - batchOrderPoster grabs everything from StackNew, posts it,
// gets BinanceId and adds it StackPosted.
type OpenOrders struct {
	Mu          sync.Mutex
	StackNew    OpenOrderStack  // [symbol]order
	StackPosted OpenOrderStack  // [symbol]order
	Busy        map[string]bool // to check if currently oms is busy with some symbol
}

type OpenOrderStack map[string]*OpenOrder

func NewOpenOrders(symbolsList []string) *OpenOrders {
	openOrders := OpenOrders{
		StackNew:    make(OpenOrderStack),
		StackPosted: make(OpenOrderStack),
		Busy:        make(map[string]bool, len(symbolsList)),
	}

	for _, symbol := range symbolsList {
		openOrders.Busy[symbol] = false
	}

	return &openOrders
}

func (oo *OpenOrders) SyncBusy(symbol string, busy bool) {
	oo.Mu.Lock()
	oo.Busy[symbol] = busy
	oo.Mu.Unlock()
}

func (oo *OpenOrders) SetBusy(symbol string, busy bool) {
	oo.Busy[symbol] = busy
}

func (oo *OpenOrders) SetMultipleBusy(openOrderStack *OpenOrderStack, busy bool) {
	for symbol := range *openOrderStack {
		oo.Busy[symbol] = busy
	}
}

// UpdateStackPosted writes just posted orders into StackPosted,
// writing to symbol-key. It releases symbol Busy to false.
func (oo *OpenOrders) UpdateStackPosted(newPostedOpenOrders *OpenOrderStack) {
	for symbolKey, newOpenOrder := range *newPostedOpenOrders {
		oo.StackPosted[symbolKey] = newOpenOrder
		oo.Busy[symbolKey] = false
	}

}

// GrabPosted grabs from StackPosted only orders, which symbols have
// new orders in StackNew.
// If order is grabbed, it means it's Binance order was canceled right after,
// so the grabbed elements are deleted from the stackPosted leaving
// up only other elements not affected.
func (oo *OpenOrders) GrabPosted(newOpenOrders *OpenOrderStack) *OpenOrderStack {
	grabbedOpenOrders := make(OpenOrderStack)

	for symbolKey, openOrder := range oo.StackPosted {
		for newOrderSymbolKey := range *newOpenOrders {
			if symbolKey == newOrderSymbolKey && openOrder.State != OpenOrderBlank {
				grabbedOpenOrders[symbolKey] = openOrder
				oo.StackPosted[symbolKey] = &OpenOrder{}
			}
		}
	}

	return &grabbedOpenOrders
}

func (oo *OpenOrders) AddNew(symbol string, ost *OrderSet) {
	// convert orderset to OpenOrder
	newOpenOrder := ost.ConvertToOpenOrder()
	// rewrite openOrder
	oo.Mu.Lock()
	oo.StackNew[symbol] = newOpenOrder
	oo.Mu.Unlock()
}

// OpenOrder
// Posted: batchOrderPoster checks if it was previously posted,
// or just added by strategy
// BinanceId: when order was previously posted and
// now should be deleted for the new one,
type OpenOrder struct {
	State      OpenOrderState
	BinanceId  int64
	Side       Side
	EntryPrice float64
	Quantity   float64
	TP         float64
	SL         float64
}

type OpenOrderState int

const (
	OpenOrderBlank OpenOrderState = iota
	OpenOrderNew
	OpenOrderPosted
)

func (oos *OpenOrderStack) ToIdList() []int64 {
	var idList []int64
	for _, openOrder := range *oos {
		idList = append(idList, openOrder.BinanceId)
	}
	return idList
}

func (oos *OpenOrderStack) ToSymbolList() []string {
	var symbolList []string
	for symbol := range *oos {
		symbolList = append(symbolList, symbol)
	}
	return symbolList
}
