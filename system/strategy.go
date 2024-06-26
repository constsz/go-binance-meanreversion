package system

import (
	"directionalMaker/data"
	"directionalMaker/oms"
	"directionalMaker/ta"
	"fmt"
	"time"
)

// StrategyList is a map[Symbol]Strategy, because one symbol can have only one strategy.
type StrategyList map[string]Strategy

// GenerateStrategyList is currently a TEMP function and later should be developed further.
func GenerateStrategyList(symbolSettings data.SymbolsSettings) *StrategyList {
	strategyList := make(StrategyList, len(symbolSettings))

	for i, s := range symbolSettings {
		var sm StrategyMode

		_, exist := strategyModes[data.SymbolNoUSDT(s.Symbol)]

		if exist {
			sm = strategyModes[data.SymbolNoUSDT(s.Symbol)]
		} else {
			fmt.Println("No StrategyMode found for ", data.SymbolNoUSDT(s.Symbol), "- using default (ModerateVolatile).")
			sm = strategyModes["DEFAULT"]
		}

		strategy := NewStrategy(
			&StrategyParams{
				TF:               sm.Timeframe,
				Quantity:         s.Quantity,
				TP:               sm.TP, // 0.5
				SL:               sm.SL, // 0.3
				LogicTP:          TP_Fixed,
				LogicSL:          SL_Fixed,
				OrderTimeSpacing: 2 * time.Minute.Milliseconds(),
			},
			&IndicatorParams{
				BetterBands: &sm.BetterBandsParams,
				TrendMove:   &sm.TrendMoveParams,
			}, s.Symbol, i)

		strategyList[s.Symbol] = *strategy
	}

	return &strategyList
}

type Strategy struct {
	// Name is automatically generated meta-name to be used for programmatic access
	Name   string
	Id     int
	Symbol string

	*StrategyParams
	*IndicatorParams
}

type StrategyParams struct {
	TF               int     // symbol specific
	Quantity         float64 // symbol specific
	TP               float64
	SL               float64
	LogicTP          LogicTP
	LogicSL          LogicSL
	OrderTimeSpacing int64
}

type IndicatorParams struct {
	TrendMove   *ta.TrendMoveParams
	BetterBands *ta.BetterBandsParams
}

type LogicTP int

const (
	TP_Fixed LogicTP = iota
	TP_ATR
)

type LogicSL int

const (
	SL_Fixed LogicSL = iota
)

func StrategyDefaultParams(tf int, quantity float64) *StrategyParams {
	return &StrategyParams{
		TF:               tf,
		Quantity:         quantity,
		TP:               0.4, // 0.5
		SL:               0.5, // 0.3
		LogicTP:          TP_Fixed,
		LogicSL:          SL_Fixed,
		OrderTimeSpacing: 5 * time.Minute.Milliseconds(),
	}
}

func NewStrategy(strategyParams *StrategyParams, indicatorParams *IndicatorParams, symbol string, strategyId int) *Strategy {
	return &Strategy{
		Name:            fmt.Sprintf("strategy-%v", strategyId),
		Id:              strategyId,
		Symbol:          symbol,
		StrategyParams:  strategyParams,
		IndicatorParams: indicatorParams,
	}
}

// Evaluate strategy
// BB: [0: Midline, 1: LowBand, 2: HiBand, 3: Regime (long/short)]
func (s *Strategy) Evaluate(c data.Candle) oms.OrderSet {
	// INITIAL
	var ost oms.OrderSet
	ost.Symbol = s.Symbol
	ost.OrderLevel.Quantity = s.Quantity
	ost.OrderTimeSpacing = s.OrderTimeSpacing

	regime := c.TA[ta.IndicatorName("BetterBands", s.Id)][3]
	TREND_UP := false
	TREND_DOWN := false
	if regime == 1.0 {
		TREND_UP = true
	} else if regime == 2.0 {
		TREND_DOWN = true
	}

	LowBand := c.TA[ta.IndicatorName("BetterBands", s.Id)][1]
	HighBand := c.TA[ta.IndicatorName("BetterBands", s.Id)][2]

	if TREND_UP {
		// ENTRY
		entryPrice := LowBand

		ost.OrderLevel.Side = oms.SideLong
		ost.OrderLevel.EntryPrice = entryPrice

		// TP
		switch s.LogicTP {
		case TP_Fixed:
			ost.OrderLevel.TP = s.TP
		}

		// STOPLOSS
		switch s.LogicSL {
		case SL_Fixed:
			ost.OrderLevel.SL = s.SL
		}
	} else if TREND_DOWN {
		// ENTRY
		entryPrice := HighBand

		ost.OrderLevel.Side = oms.SideShort
		ost.OrderLevel.EntryPrice = entryPrice

		// TP
		switch s.LogicTP {
		case TP_Fixed:
			ost.OrderLevel.TP = s.TP
		}

		// STOPLOSS
		switch s.LogicSL {
		case SL_Fixed:
			ost.OrderLevel.SL = s.SL
		}
	}

	return ost
}
