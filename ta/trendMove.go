package ta

import (
	"directionalMaker/data"
	"fmt"
)

// TrendMove indicator stores the following values in candle.TA[tm.Name]:
// 0: TrendLine
// 1: MoveLine
// 2: TrendReturns
// 3: MoveReturns
// 4: TrendReturnsEma
// 5: MoveReturnsEma
type TrendMove struct {
	Name string
	// StrategyId is used to generate Name, like so: "bb-12" and "tm-12"
	StrategyId int
	// Window used to cut NA rows of df after preload calc
	Window int
	// Type used for Chart frontend
	Type string

	Params TrendMoveParams
}

func (tm *TrendMove) GetParams() []string {
	indParams := make([]string, 2)
	indParams[0] = tm.Name
	indParams[1] = tm.Type

	return indParams
}

type TrendMoveParams struct {
	TrendPeriod    int
	MovePeriod     int
	TrendSmoothing int
	MoveSmoothing  int
}

func TrendMoveDefaultParams() *TrendMoveParams {
	return &TrendMoveParams{
		TrendPeriod:    400, // 120
		MovePeriod:     60,  // 60?
		TrendSmoothing: 42,  // 21
		MoveSmoothing:  28,  // 14
	}
}

func NewTrendMove(strategyId int, trendMoveParams *TrendMoveParams) *TrendMove {
	return &TrendMove{
		Name:       fmt.Sprintf("tm-%d", strategyId),
		StrategyId: strategyId,
		Window:     trendMoveParams.TrendPeriod + trendMoveParams.TrendSmoothing + 1,
		Type:       "TrendMove",
		Params:     *trendMoveParams,
	}
}

// InitialCalc stores the following values in candles[i].TA[tm.Name]:
// 0: TrendLine
// 1: MoveLine
// 2: TrendReturns
// 3: MoveReturns
// 4: TrendReturnsEma
// 5: MoveReturnsEma
func (tm *TrendMove) InitialCalc(candles *data.Candles) {
	var prevTrend float64
	var prevMove float64
	var prevTrendReturnsEma float64
	var prevMoveReturnsEma float64

	for i := range *candles {
		if i > tm.Params.TrendPeriod {
			// ----------------------------------------------------------------------
			// TREND Line      [0: TrendLine]
			// Prepare a sample-window of data for Trend EMA
			candlesSampleTrend := (*candles)[i-tm.Params.TrendPeriod+1 : i+1]

			// Include only required values
			pricesTrend := make([]float64, tm.Params.TrendPeriod)
			for j, c := range candlesSampleTrend {
				pricesTrend[j] = c.Close
			}

			// Calculate Trend line using EMA
			trendVal := CalcEMA(&pricesTrend, prevTrend)
			(*candles)[i].TA[tm.Name] = append((*candles)[i].TA[tm.Name], trendVal)

			prevTrend = trendVal

			// ----------------------------------------------------------------------
			// MOVE Line        [1: MoveLine]
			// Prepare a sample-window of data for Move EMA
			candlesSampleMove := (*candles)[i-tm.Params.MovePeriod+1 : i+1]

			// Include only required values
			pricesMove := make([]float64, tm.Params.MovePeriod)
			for j, c := range candlesSampleMove {
				pricesMove[j] = c.Close
			}

			// Calculate Move line using EMA
			moveVal := CalcEMA(&pricesMove, prevMove)
			(*candles)[i].TA[tm.Name] = append((*candles)[i].TA[tm.Name], moveVal)

			prevMove = moveVal

		}
		// ----------------------------------------------------------------------
		// RETURNS for Trend and Move lines
		// [2: TrendReturns]
		if i > tm.Params.TrendPeriod+1 {
			if (*candles)[i-1].TA[tm.Name][0] > 0 {
				trendReturns := (*candles)[i].TA[tm.Name][0] / (*candles)[i-1].TA[tm.Name][0]
				(*candles)[i].TA[tm.Name] = append((*candles)[i].TA[tm.Name], trendReturns)
			}
			// [3: MoveReturns]
			if (*candles)[i-1].TA[tm.Name][1] > 0 {
				moveReturns := (*candles)[i].TA[tm.Name][1] / (*candles)[i-1].TA[tm.Name][1]
				(*candles)[i].TA[tm.Name] = append((*candles)[i].TA[tm.Name], moveReturns)
			}
		}
		// ----------------------------------------------------------------------
		// RETURNS EMA
		if i > tm.Params.TrendPeriod+tm.Params.TrendSmoothing {
			//TrendReturnsEma
			//[4: TrendReturnsEma]

			//Prepare a sample-window of data for Trend EMA
			candlesSampleTrendReturns := (*candles)[i-tm.Params.TrendSmoothing+1 : i+1]

			// Include only required values
			valuesTrendReturns := make([]float64, tm.Params.TrendSmoothing)
			for j, c := range candlesSampleTrendReturns {
				valuesTrendReturns[j] = c.TA[tm.Name][2]
			}

			// Calculate Trend line using EMA
			trendReturnsEma := CalcEMA(&valuesTrendReturns, prevTrendReturnsEma)
			(*candles)[i].TA[tm.Name] = append((*candles)[i].TA[tm.Name], trendReturnsEma)

			prevTrendReturnsEma = trendReturnsEma

			//MoveReturnsEma
			//[5: MoveReturnsEma]
			//Prepare a sample-window of data for Trend EMA
			candlesSampleMoveReturns := (*candles)[i-tm.Params.MoveSmoothing+1 : i+1]

			// Include only required values
			valuesMoveReturns := make([]float64, tm.Params.MoveSmoothing)
			for j, c := range candlesSampleMoveReturns {
				valuesMoveReturns[j] = c.TA[tm.Name][2]
			}

			// Calculate Move line using EMA
			moveReturnsEma := CalcEMA(&valuesMoveReturns, prevMoveReturnsEma)
			(*candles)[i].TA[tm.Name] = append((*candles)[i].TA[tm.Name], moveReturnsEma)

			prevMoveReturnsEma = moveReturnsEma

		}
	}
}

// CandleCalc calculates indicator value for a single candle
// 0: TrendLine
// 1: MoveLine
// 2: TrendReturns
// 3: MoveReturns
// 4: TrendReturnsEma
// 5: MoveReturnsEma
func (tm *TrendMove) CandleCalc(candles *data.Candles) {
	last := len(*candles) - 1
	prev := last - 1

	// ----------------------------------------------------------------------
	// TREND Line      [0: TrendLine]
	// Prepare a sample-window of data for Trend EMA
	candlesSampleTrend := (*candles)[last-tm.Params.TrendPeriod+1:]

	// Include only required values
	closePrices := make([]float64, tm.Params.TrendPeriod)
	for j, c := range candlesSampleTrend {
		closePrices[j] = c.Close
	}

	// Calculate Trend line using EMA
	trendVal := CalcEMA(&closePrices, (*candles)[prev].TA[tm.Name][0])
	(*candles)[last].TA[tm.Name] = append((*candles)[last].TA[tm.Name], trendVal)

	// ----------------------------------------------------------------------
	// MOVE Line        [1: MoveLine]
	// Prepare a sample-window of data for Move EMA
	candlesSampleMove := (*candles)[last-tm.Params.MovePeriod+1:]

	// Include only required values
	closePricesMove := make([]float64, tm.Params.MovePeriod)
	for j, c := range candlesSampleMove {
		closePricesMove[j] = c.Close
	}

	// Calculate Move line using EMA
	moveVal := CalcEMA(&closePricesMove, (*candles)[prev].TA[tm.Name][1])
	(*candles)[last].TA[tm.Name] = append((*candles)[last].TA[tm.Name], moveVal)

	// ----------------------------------------------------------------------
	// RETURNS for Trend and Move lines
	// [2: TrendReturns]
	if (*candles)[prev].TA[tm.Name][0] > 0 {
		trendReturns := (*candles)[last].TA[tm.Name][0] / (*candles)[prev].TA[tm.Name][0]
		(*candles)[last].TA[tm.Name] = append((*candles)[last].TA[tm.Name], trendReturns)
	}
	// [3: MoveReturns]
	if (*candles)[prev].TA[tm.Name][1] > 0 {
		moveReturns := (*candles)[last].TA[tm.Name][1] / (*candles)[prev].TA[tm.Name][1]
		(*candles)[last].TA[tm.Name] = append((*candles)[last].TA[tm.Name], moveReturns)
	}

	// ----------------------------------------------------------------------
	// RETURNS EMA

	// TrendReturnsEma
	// [4: TrendReturnsEma]

	//Prepare a sample-window of data for Trend EMA
	candlesSampleTrendReturns := (*candles)[last-tm.Params.TrendSmoothing+1:]

	// Include only required values
	valuesTrendReturns := make([]float64, tm.Params.TrendSmoothing)
	for j, c := range candlesSampleTrendReturns {
		valuesTrendReturns[j] = c.TA[tm.Name][2]
	}

	// Calculate Trend line using EMA
	trendReturnsEma := CalcEMA(&valuesTrendReturns, (*candles)[prev].TA[tm.Name][4])
	(*candles)[last].TA[tm.Name] = append((*candles)[last].TA[tm.Name], trendReturnsEma)

	//MoveReturnsEma
	//[5: MoveReturnsEma]
	//Prepare a sample-window of data for Trend EMA
	candlesSampleMoveReturns := (*candles)[last-tm.Params.MoveSmoothing+1:]

	// Include only required values
	valuesMoveReturns := make([]float64, tm.Params.MoveSmoothing)
	for j, c := range candlesSampleMoveReturns {
		valuesMoveReturns[j] = c.TA[tm.Name][2]
	}

	// Calculate Move line using EMA
	moveReturnsEma := CalcEMA(&valuesMoveReturns, (*candles)[prev].TA[tm.Name][5])
	(*candles)[last].TA[tm.Name] = append((*candles)[last].TA[tm.Name], moveReturnsEma)

}
