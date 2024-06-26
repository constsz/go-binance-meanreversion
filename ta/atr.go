package ta

import (
	"directionalMaker/data"
	"fmt"
	"math"
)

type ATR struct {
	Name string
	// StrategyId is used to generate Name, like so: "bb-12" and "tm-12"
	StrategyId int
	Period     int
	Type       string
}

// NewATR default period is 30
// Values: [0: ATR, 1: TrueRange]
func NewATR(strategyId int, period int) *ATR {
	return &ATR{
		Name:       fmt.Sprintf("atr-%d", strategyId),
		StrategyId: strategyId,
		Period:     period,
		Type:       "ATR",
	}
}

func (atr *ATR) InitialCalc(candles *data.Candles) {
	// Loop over candles
	for i := range *candles {
		if i > 0 {
			// ------------------------------------
			// TrueRange   			[1: TrueRange]
			// Create empty slice for TA[atr]
			taAtrSlice := make([]float64, 2)

			// Calculate TrueRange for current candle
			trueRange := TrueRange((*candles)[i-1 : i+1])

			// Save TrueRange into c.TA[atr.Name][1], because [0] goes to ATR
			taAtrSlice[1] = trueRange

			// Save ATR values (first iteration of ATR will have 0 value)
			(*candles)[i].TA[atr.Name] = append((*candles)[i].TA[atr.Name], taAtrSlice...)

			// ------------------------------------
			// ATR 					[0: ATR]
			// i.e. calculate RMA on TrueRange series
			//// Output: one number (ATR for period)
			if i > atr.Period {
				// Extract TrueRange in a series values from candles
				trueRangeSeries := make([]float64, atr.Period)

				// From a atr.Period extract TrueRange values into a series
				for j, c := range (*candles)[i-atr.Period+1 : i+1] {
					trueRangeSeries[j] = c.TA[atr.Name][1]
				}

				// Calculate ATR
				atrVal := CalcRMA(&trueRangeSeries, (*candles)[i-1].TA[atr.Name][0])
				(*candles)[i].TA[atr.Name][0] = atrVal

			}

		}

	}
}

// TrueRange receives 2 candles: last and previous. It use prev. to calc last.
func TrueRange(candles data.Candles) float64 {
	c := candles[1]
	cp := candles[0]

	high := c.High
	low := c.Low
	prevClose := cp.Close

	valueCandidates := []float64{
		high - low,
		math.Abs(high - prevClose),
		math.Abs(low - prevClose),
	}

	return Max(valueCandidates)
}

func (atr *ATR) CandleCalc(candles *data.Candles) {
	last := len(*candles) - 1
	prev := last - 1

	// ------------------------------------
	// TrueRange   			[1: TrueRange]
	// Create empty slice for TA[atr]
	taAtrSlice := make([]float64, 2)

	// Calculate TrueRange for current candle
	trueRange := TrueRange((*candles)[prev:])

	// Save TrueRange into c.TA[atr.Name][1], because [0] goes to ATR
	taAtrSlice[1] = trueRange

	// Save ATR values (first iteration of ATR will have 0 value)
	(*candles)[last].TA[atr.Name] = append((*candles)[last].TA[atr.Name], taAtrSlice...)

	// ------------------------------------
	// ATR 					[0: ATR]
	// i.e. calculate RMA on TrueRange series
	//// Output: one number (ATR for period)
	// Extract TrueRange in a series values from candles
	trueRangeSeries := make([]float64, atr.Period)

	// From a atr.Period extract TrueRange values into a series
	for j, c := range (*candles)[last-atr.Period+1:] {
		trueRangeSeries[j] = c.TA[atr.Name][1]
	}

	// Calculate ATR
	atrVal := CalcRMA(&trueRangeSeries, (*candles)[prev].TA[atr.Name][0])
	(*candles)[last].TA[atr.Name][0] = atrVal

}

/* UNUSED */
// Calc ATR: calculate RMA on TrueRange series
// Output: one number (ATR for period)
//func (atr *ATR) Calc(sampleOfCandles *data.Candles, prevRma float64) float64 {
//	// Extract TrueRange in a series values from candles
//	trueRangeSeries := make([]float64, atr.Period)
//
//	for i, c := range *sampleOfCandles {
//		trueRangeSeries[i] = c.TA[atr.Name][1]
//	}
//
//	atrVal := CalcRMA(&trueRangeSeries, prevRma)
//
//	return atrVal
//}
