package oms

import (
	"directionalMaker/data"
	"sync"
)

type Inventory struct {
	Capacity int
	Mu       sync.Mutex
	Stack    map[string]*SymbolInventory
}

// SymbolInventory if capacity=1, and there is already 1 active trade,
// OMS will skip and not add new orders.
type SymbolInventory struct {
	Capacity           int
	ActiveTrades       []*ActiveTrade
	LastTimeOrderAdded int64
}

func (inv *Inventory) InsertActiveTradeSync(symbol string, newActiveTrade *ActiveTrade) {
	inv.Mu.Lock()
	inv.Stack[symbol].ActiveTrades = append(inv.Stack[symbol].ActiveTrades, newActiveTrade)
	inv.Mu.Unlock()
}

func (inv *Inventory) InsertActiveTrade(symbol string, newActiveTrade *ActiveTrade) {
	inv.Stack[symbol].ActiveTrades = append(inv.Stack[symbol].ActiveTrades, newActiveTrade)
}

// NewInventory - for default capacity use value 0
func NewInventory(symbolsSettings *data.SymbolsSettings, capacity int) *Inventory {
	defaultCapacity := 1
	if capacity == 0 {
		capacity = defaultCapacity
	}

	inventory := Inventory{
		Capacity: capacity,
		Stack:    make(map[string]*SymbolInventory, len(*symbolsSettings)),
	}

	for _, symSetting := range *symbolsSettings {
		inventory.Stack[symSetting.Symbol] = &SymbolInventory{}
	}

	return &inventory
}

type ActiveTrade struct {
	Side       Side
	EntryPrice float64
	EntryTime  int64
	TP         float64
	SL         float64
	//BinanceTradeId int64
	Quantity       float64
	BinanceIdEntry int64
	BinanceIdTP    int64
	BinanceIdSL    int64
}

// RemoveActiveOrder is a Backtester function, can be ignored for now.
func (si *SymbolInventory) RemoveActiveOrder(i int) {
	//fmt.Println("Removing Active Order...")
	//fmt.Println("LEN Before:", len(si.ActiveTrades))

	si.ActiveTrades[i] = si.ActiveTrades[len(si.ActiveTrades)-1]
	si.ActiveTrades = si.ActiveTrades[:len(si.ActiveTrades)-1]

	//fmt.Println("LEN After:", len(si.ActiveTrades))
	//if len(si.ActiveTrades) > 1 {
	//
	//} else if len(si.ActiveTrades) == 1 {
	//	si.ActiveTrades = []ActiveTrade{}
	//}
}
