package ta

import (
	"directionalMaker/data"
	"fmt"
	"math"
)

// (!) Regime: 0 = wait, 1 = long, 2 = short

// BetterBands works only after TrendMove is calculated.
// It directly calls candle.TA["TrendMoveName"] from Calc function.
// Values:  [0: Midline, 1: LowBand, 2: HiBand, 3: Regime (long/short)]
type BetterBands struct {
	// Name can be used in future to use multiple instances of this indicator
	// in a single strategy (for example, for 2 timeframe settings)
	Name string
	// StrategyId is used to generate Name, like so: "bb-12" and "tm-12"
	StrategyId int
	// Window used to cut NA rows of df after preload calc
	Window int
	// Type used for Chart frontend
	Type string

	Params BetterBandsParams
}

func (bb *BetterBands) GetParams() []string {
	indParams := make([]string, 2)
	indParams[0] = bb.Name
	indParams[1] = bb.Type

	return indParams
}

type BetterBandsParams struct {
	BandsPeriod        int
	ATRMultiplier      float64
	BandMult2          float64
	BandMult3          float64
	MoveBoostParameter float64
	MidlineTrendParam  float64
	MidlineBoostParam  float64
}

// TODO : DEBUG
var DEBUG bool = false
var minBandRange = 1.0025 // 1.005

func BetterBandsDefaultParams() *BetterBandsParams {
	return &BetterBandsParams{
		BandsPeriod:        10,
		ATRMultiplier:      2,
		BandMult2:          1.2, // 3
		BandMult3:          1.2, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.6, // 2
		MidlineBoostParam: 1.3, // 2
	}
}

func NewBetterBands(strategyId int, betterBandsParams *BetterBandsParams) *BetterBands {
	return &BetterBands{
		Name:       fmt.Sprintf("bb-%d", strategyId),
		StrategyId: strategyId,
		Window:     betterBandsParams.BandsPeriod,
		Type:       "BetterBands",
		Params:     *betterBandsParams,
	}
}

// InitialCalc calculates preloaded historical candles using Calc function.
func (bb *BetterBands) InitialCalc(candles *data.Candles) {
	var prevMidline float64
	//trendMove := fmt.Sprintf("tm-%d", bb.StrategyId)

	// Loop over Candles
	for i := range *candles {
		// Start calculation when enough candles available
		if i >= bb.Window /*&& len((*candles)[i].TA[trendMove]) == 6*/ {
			// ----------------------------------------------------------------------
			// M I D L I N E       [0: Midline]
			// Prepare a sample-window of data for Midline EMA
			candlesSampleMidline := (*candles)[i-bb.Params.BandsPeriod+1 : i+1]

			// Include only required values
			closePrices := make([]float64, bb.Params.BandsPeriod)
			for j, c := range candlesSampleMidline {
				closePrices[j] = c.Close
			}

			// Calculate Midline using EMA
			midline := CalcEMA(&closePrices, prevMidline)
			(*candles)[i].TA[bb.Name] = append((*candles)[i].TA[bb.Name], midline)

			prevMidline = midline

			// ----------------------------------------------------------------------
			// BETTER BANDS       [1: HiBand, 2: LowBand, 3: Regime (long/short)]
			// Prepare a sample-window of data for Midline EMA
			atrName := fmt.Sprintf("atr-%d", bb.StrategyId)

			// ATR and bandPct
			atr := (*candles)[i].TA[atrName][0]
			atrPct := (atr / (*candles)[i].Close) + 1
			bandPct := (atrPct-1)*bb.Params.ATRMultiplier + 1

			// BetterBands
			bbValues := bb.Calc((*candles)[i], bandPct)
			(*candles)[i].TA[bb.Name] = append((*candles)[i].TA[bb.Name], bbValues...)
		}
	}
}

// Calc calculates BetterBands values: [1: LowBand, 2: HighBand, 3: Regime (long/short)]
func (bb *BetterBands) Calc(candle data.Candle, bandPct float64) []float64 {

	trendMove := fmt.Sprintf("tm-%d", bb.StrategyId)

	midLine := candle.TA[bb.Name][0]

	//trendLine := candle.TA[trendMove][0]
	//moveLine := candle.TA[trendMove][1]
	//trendReturns := candle.TA[trendMove][2]
	moveReturns := candle.TA[trendMove][3]
	trendReturnsEma := candle.TA[trendMove][4]
	moveReturnsEma := candle.TA[trendMove][5]

	// ----------------------------------------------------
	// LOGIC

	TREND_UP := false
	if trendReturnsEma > 1 {
		TREND_UP = true
	}
	TREND_DOWN := false
	if trendReturnsEma < 1 {
		TREND_DOWN = true
	}
	MOVE_UP := false
	if moveReturnsEma > 1 {
		MOVE_UP = true
	}
	MOVE_DOWN := false
	if moveReturnsEma < 1 {
		MOVE_DOWN = true
	}
	MOVE_BOOST := false
	if math.Abs(moveReturns-1) > bb.Params.MoveBoostParameter {
		MOVE_BOOST = true
	}

	// MidLine Adjustments
	MidlineTrendAdjustment := ((bandPct - 1) / bb.Params.MidlineTrendParam) + 1
	MidlineBoostAdjustment := ((bandPct - 1) / bb.Params.MidlineBoostParam) + 1

	// PCT for Bands l2,l3
	bandPct_2 := (bandPct-1)*bb.Params.BandMult2 + 1
	bandPct_3 := (bandPct-1)*bb.Params.BandMult3 + 1

	bandPct_main := bandPct_2

	if TREND_UP && MOVE_UP {
		midLine = midLine * MidlineTrendAdjustment
	}
	if TREND_DOWN && MOVE_DOWN {
		midLine = midLine / MidlineTrendAdjustment
	}

	if TREND_UP && MOVE_UP {
		bandPct_main = bandPct
	}
	if TREND_DOWN && MOVE_DOWN {
		bandPct_main = bandPct
	}

	if TREND_UP && MOVE_DOWN && MOVE_BOOST {
		bandPct_main = bandPct_3
	}
	if TREND_DOWN && MOVE_UP && MOVE_BOOST {
		bandPct_main = bandPct_3
	}

	if TREND_UP && MOVE_UP && MOVE_BOOST {
		midLine = midLine * MidlineBoostAdjustment
	}
	if TREND_DOWN && MOVE_DOWN && MOVE_BOOST {
		midLine = midLine / MidlineBoostAdjustment
	}

	if bandPct_main < minBandRange {
		bandPct_main = minBandRange
	}

	// TODO : DEBUG -> must be COMMENTED
	if bandPct_main > 1.05 {
		bandPct_main = 1.05
	}

	// Final Values for Bands
	l1 := midLine / bandPct_main
	h1 := midLine * bandPct_main

	trend := 0.0
	if TREND_UP {
		trend = 1.0
	} else if TREND_DOWN {
		trend = 2.0
	}

	bbValues := []float64{l1, h1, trend}

	return bbValues

}

func (bb *BetterBands) CandleCalc(candles *data.Candles) {
	last := len(*candles) - 1
	prev := last - 1

	// ----------------------------------------------------------------------
	// M I D L I N E       [0: Midline]
	// Prepare a sample-window of data for Midline EMA
	candlesSampleMidline := (*candles)[last-bb.Params.BandsPeriod+1:]

	// Include only required values
	closePrices := make([]float64, bb.Params.BandsPeriod)
	for j, c := range candlesSampleMidline {
		closePrices[j] = c.Close
	}

	// Calculate Midline using EMA
	midline := CalcEMA(&closePrices, (*candles)[prev].TA[bb.Name][0])
	(*candles)[last].TA[bb.Name] = append((*candles)[last].TA[bb.Name], midline)

	// ----------------------------------------------------------------------
	// BETTER BANDS       [1: HiBand, 2: LowBand, 3: Regime (long/short)]
	// Prepare a sample-window of data for Midline EMA
	atrName := fmt.Sprintf("atr-%d", bb.StrategyId)

	// ATR and bandPct
	atr := (*candles)[last].TA[atrName][0]
	atrPct := (atr / (*candles)[last].Close) + 1
	bandPct := (atrPct-1)*bb.Params.ATRMultiplier + 1

	// TODO : DEBUG -> must be COMMENTED
	if DEBUG {
		bandPct = 1.001
	}

	// BetterBands
	bbValues := bb.Calc((*candles)[last], bandPct)
	(*candles)[last].TA[bb.Name] = append((*candles)[last].TA[bb.Name], bbValues...)

}
