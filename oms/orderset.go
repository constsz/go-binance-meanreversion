package oms

import "fmt"

type OrderSet struct {
	Symbol           string
	InstantAction    InstantAction
	OrderLevel       OrderLevel
	OrderTimeSpacing int64
}

func (ost *OrderSet) ConvertToOpenOrder() *OpenOrder {
	ol := ost.OrderLevel
	return &OpenOrder{
		State:      OpenOrderNew,
		Side:       ol.Side,
		EntryPrice: ol.EntryPrice,
		Quantity:   ol.Quantity,
		TP:         ol.TP,
		SL:         ol.SL,
	}
}

type OrderLevel struct {
	Side       Side
	EntryPrice float64
	Quantity   float64
	TP         float64
	SL         float64
}

func (ost *OrderSet) Print() {
	fmt.Println("\n---------------")
	fmt.Println("OrderSet")
	fmt.Println("Symbol", ost.Symbol)
	//fmt.Println("InstantAction:", ost.InstantAction)
	fmt.Println()
	fmt.Println("OrderLevel:")
	fmt.Println("Side      :", ost.OrderLevel.Side)
	fmt.Println("EntryPrice:", ost.OrderLevel.EntryPrice)
	fmt.Println("TP        :", ost.OrderLevel.TP)
	fmt.Println("SL        :", ost.OrderLevel.SL)
	fmt.Println("Quantity  :", ost.OrderLevel.Quantity)
	fmt.Println()
	fmt.Println("OrderTimeSpacing:", ost.OrderTimeSpacing)
}

type Side int

const (
	SideNone Side = iota
	SideLong
	SideShort
)

type InstantAction struct {
	Type InstantActionType
	Info []any
}

type InstantActionType int

const (
	IA_None InstantActionType = iota
	IA_ExitLongs
	IA_ExitShorts
	IA_ExitLongs_EntryGreaterThan
	IA_ExitShorts_EntryGreaterThan
	IA_ExitByTTL
)

type Trades []Trade

type Trade struct {
	Side       Side
	EntryTime  int64
	ExitTime   int64
	Quantity   int64
	EntryPrice float64
	ExitPrice  float64
}
