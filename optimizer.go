package main

import (
	"directionalMaker/data"
	"directionalMaker/oms"
	"directionalMaker/system"
	"directionalMaker/ta"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"sort"
	"time"
)

// Пишем синхронный код, и запускаем несколько копий отдельно - так быстрее и проще написать.

func Optimizer() {
	cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println()

	// List of symbols
	//symbols := data.SymbolUSDTList([]string{
	//	"CFX", "MASK", "OP", "SUSHI", "FIL", "ACH", "HOOK", "XEM", "MKR", "API3",
	//	"COCOS", "REN", "TOMO", "MAGIC", "LDO", "BNB", "FTM", "DYDX", "ANKR", "DOT", "MANA",
	//	"DOGE", "1000SHIB", "XRP", "ADA", "AVAX", "YFI", "BNX", "KNC", "QTUM",
	//	"XLM", "LRC", "C98", "BLZ", "STORJ", "SXP", "ALICE", "ALGO", "ONE", "EOS", "CRV",
	//	"ARPA", "AGIX", "T", "GALA", "SNX", "HIGH", "ONT", "GMT", "LINA", "APE", "AXS",
	//})

	symbols := data.SymbolUSDTList([]string{
		"ARB", "XRP", "EOS",
	})

	var resultsMinWinRate float64 = 77
	var resultsMinPnl float64 = 0
	var resultsMinTrades int = 1

	// --------------------------------------------------------------------------------

	var (
		apiKey    = "CZYRAfunS6FVDQJH5cjZCQsA89SsiJrVqCN5G5DR8HGzyiVJXPw2k4ASChjR6Lgk"
		secretKey = "lGvXuV3iaYDo4LovFu03nOPCKguCFo83M2uZpVxRe9ANGcsPsyo0akZJOP7mVsKS"
	)
	client := binance.NewFuturesClient(apiKey, secretKey)
	hoursWindow := 8

	timeTotal := time.Now()
	optimiserResults := make(OptimiserResults)

	totalNum := calculateTotalNumberOfParameterVariations(pvStrategy, pvBetterBands, pvTrendMove)

	fmt.Println("Total number of parameter variations:", totalNum)

	// -----------------------------------------------------------------------------
	// LOAD DATA
	for _, symbol := range symbols {
		timeSymbolDownload := time.Now()

		// Load Trades data for the last 12 Hours and convert it to tick candles
		var candles data.Candles
		candles.LoadHistoricalData(client, symbol, 90, hoursWindow)

		//filePath := "D:\\programming\\Projects_Trading\\candles\\main\\candles_200\\ADAUSDT\\ADAUSDT-candles-2023-03-21.csv"
		//candles := data.LoadCSVCandles(filePath)

		fmt.Println()
		fmt.Println(symbol, "data loaded | Minutes elapsed:", time.Since(timeSymbolDownload).Minutes())

		fmt.Println()
		fmt.Println("-----------------------------------")
		fmt.Println()

		// -----------------------------------------------------------------------------
		// BACKTESTING PROCESS
		timeSymbolBacktesting := time.Now()
		var smCounter int
		for _, xTrendPeriod := range pvTrendMove["TrendPeriod"] {
			for _, xMovePeriod := range pvTrendMove["MovePeriod"] {
				for _, xTrendSmoothing := range pvTrendMove["TrendSmoothing"] {
					for _, xMoveSmoothing := range pvTrendMove["MoveSmoothing"] {
						for _, xBandsPeriod := range pvBetterBands["BandsPeriod"] {
							for _, xATRMultiplier := range pvBetterBands["ATRMultiplier"] {
								for _, xBandMult2 := range pvBetterBands["BandMult2"] {
									for _, xBandMult3 := range pvBetterBands["BandMult3"] {
										for _, xMoveBoostParameter := range pvBetterBands["MoveBoostParameter"] {
											for _, xMidlineTrendParam := range pvBetterBands["MidlineTrendParam"] {
												for _, xMidlineBoostParam := range pvBetterBands["MidlineBoostParam"] {
													for _, xTP := range pvStrategy["TP"] {
														for _, xSL := range pvStrategy["SL"] {

															trendMoveParams := ta.TrendMoveParams{
																TrendPeriod:    int(xTrendPeriod),
																MovePeriod:     int(xMovePeriod),
																TrendSmoothing: int(xTrendSmoothing),
																MoveSmoothing:  int(xMoveSmoothing),
															}

															betterBandsParams := ta.BetterBandsParams{
																BandsPeriod:        int(xBandsPeriod),
																ATRMultiplier:      xATRMultiplier,
																BandMult2:          xBandMult2,
																BandMult3:          xBandMult3,
																MoveBoostParameter: xMoveBoostParameter,
																MidlineTrendParam:  xMidlineTrendParam,
																MidlineBoostParam:  xMidlineBoostParam,
															}

															sm := system.StrategyMode{
																Name:              data.SymbolNoUSDT(symbol),
																TP:                xTP,
																SL:                xSL,
																BetterBandsParams: betterBandsParams,
																TrendMoveParams:   trendMoveParams,
																Timeframe:         20,
																QtyUsd:            100,
															}

															candlesCopy := make(data.Candles, len(candles))
															copy(candlesCopy, candles)

															// Initialize symbol indicators
															tm := ta.NewTrendMove(smCounter, &sm.TrendMoveParams)
															atr := ta.NewATR(smCounter, 60)
															bb := ta.NewBetterBands(smCounter, &sm.BetterBandsParams)

															// Calculate indicators
															tm.InitialCalc(&candlesCopy)
															atr.InitialCalc(&candlesCopy)

															candlesCopy.CutUnused(tm.Window - bb.Window)
															bb.InitialCalc(&candlesCopy)
															candlesCopy.CutUnused(bb.Window)

															strategy := *system.NewStrategy(
																&system.StrategyParams{
																	TF:               sm.Timeframe,
																	Quantity:         sm.QtyUsd,
																	TP:               sm.TP, // 0.5
																	SL:               sm.SL, // 0.3
																	LogicTP:          system.TP_Fixed,
																	LogicSL:          system.SL_Fixed,
																	OrderTimeSpacing: 1 * time.Second.Milliseconds(),
																},
																&system.IndicatorParams{
																	BetterBands: &sm.BetterBandsParams,
																	TrendMove:   &sm.TrendMoveParams,
																}, symbol, smCounter)

															strategy.Name = fmt.Sprintf("%s-%v", sm.Name, smCounter)

															// Run optimiser's Backtest process and return results
															strategyModeResults := backtestStrategyMode(&candlesCopy, strategy)

															// Append results to global results for this symbol
															if strategyModeResults.PnL > resultsMinPnl && strategyModeResults.Trades > resultsMinTrades && strategyModeResults.WinRate > resultsMinWinRate {
																optimiserResults[symbol] = append(optimiserResults[symbol], strategyModeResults)
															}

															smCounter++

														}
													}
												}
											}
										}
									}
									fmt.Printf("%s | %d%s | Minutes elapsed: %.2f", symbol, int(float64(smCounter)/(float64(totalNum)/100)), "%", time.Since(timeSymbolBacktesting).Minutes())
									fmt.Println("\033[1A\033[K")
								}
								runtime.GC()
								debug.FreeOSMemory()
							}
						}
					}
				}
			}
		}

		if len(optimiserResults[symbol]) > 0 {
			// Final Results Output
			fmt.Println()
			fmt.Println()

			//Sort OptimiserResults by Pnl and number of trades
			sort.Slice(optimiserResults[symbol], func(i, j int) bool {
				return optimiserResults[symbol][i].PnL > optimiserResults[symbol][j].PnL
			})

			fmt.Println()
			fmt.Println("===================================")
			fmt.Println("          by PnL")
			fmt.Println()
			fmt.Println(data.SymbolNoUSDT(symbol))

			for i, result := range optimiserResults[symbol] {
				if i < 9 {
					fmt.Println()
					fmt.Println("-----------------------------------")
					fmt.Println(result.IterationName)
					fmt.Println("PnL:             ", fmt.Sprintf("%.2f%s", result.PnL, "%"))
					fmt.Println("Balance:         ", fmt.Sprintf("%.2f", result.Balance))
					fmt.Println("MaxDrawdown:     ", fmt.Sprintf("%.2f", result.MaxDrawdown))
					fmt.Println("WinRate:         ", fmt.Sprintf("%.2f%s", result.WinRate, "%"))
					fmt.Println("Trades:          ", result.Trades)
					fmt.Println()
					fmt.Println("TP, SL: ", result.StrategyParams.TP, result.StrategyParams.SL)
					fmt.Println("TrendMove:")
					fmt.Println("-- TrendPeriod:       ", result.IndicatorParams.TrendMove.TrendPeriod)
					fmt.Println("-- MovePeriod:        ", result.IndicatorParams.TrendMove.MovePeriod)
					fmt.Println("-- TrendSmoothing:       ", result.IndicatorParams.TrendMove.TrendSmoothing)
					fmt.Println("-- MoveSmoothing:        ", result.IndicatorParams.TrendMove.MoveSmoothing)
					fmt.Println("BetterBands:")
					fmt.Println("-- BandsPeriod:       ", result.IndicatorParams.BetterBands.BandsPeriod)
					fmt.Println("-- ATRMultiplier:     ", result.IndicatorParams.BetterBands.ATRMultiplier)
					fmt.Println("-- BandMult2:         ", result.IndicatorParams.BetterBands.BandMult2)
					fmt.Println("-- BandMult3:         ", result.IndicatorParams.BetterBands.BandMult3)
					fmt.Println("-- MoveBoostParameter:", result.IndicatorParams.BetterBands.MoveBoostParameter)
					fmt.Println("-- MidlineTrendParam: ", result.IndicatorParams.BetterBands.MidlineTrendParam)
					fmt.Println("-- MidlineBoostParam: ", result.IndicatorParams.BetterBands.MidlineBoostParam)
				}
			}

			//Sort OptimiserResults by Pnl and number of trades
			sort.Slice(optimiserResults[symbol], func(i, j int) bool {
				return optimiserResults[symbol][i].WinRate > optimiserResults[symbol][j].WinRate
			})

			fmt.Println()
			fmt.Println("-----------------------------------")
			fmt.Println("          by WIN RATE")
			fmt.Println()
			fmt.Println(data.SymbolNoUSDT(symbol))

			for i, result := range optimiserResults[symbol] {
				if i < 9 {
					fmt.Println()
					fmt.Println("-----------------------------------")
					fmt.Println(result.IterationName)
					fmt.Println("PnL:             ", fmt.Sprintf("%.2f%s", result.PnL, "%"))
					fmt.Println("Balance:         ", fmt.Sprintf("%.2f", result.Balance))
					fmt.Println("MaxDrawdown:     ", fmt.Sprintf("%.2f", result.MaxDrawdown))
					fmt.Println("WinRate:         ", fmt.Sprintf("%.2f%s", result.WinRate, "%"))
					fmt.Println("Trades:          ", result.Trades)
					fmt.Println()
					fmt.Println("TP, SL: ", result.StrategyParams.TP, result.StrategyParams.SL)
					fmt.Println("TrendMove:")
					fmt.Println("-- TrendPeriod:       ", result.IndicatorParams.TrendMove.TrendPeriod)
					fmt.Println("-- MovePeriod:        ", result.IndicatorParams.TrendMove.MovePeriod)
					fmt.Println("-- TrendSmoothing:       ", result.IndicatorParams.TrendMove.TrendSmoothing)
					fmt.Println("-- MoveSmoothing:        ", result.IndicatorParams.TrendMove.MoveSmoothing)
					fmt.Println("BetterBands:")
					fmt.Println("-- BandsPeriod:       ", result.IndicatorParams.BetterBands.BandsPeriod)
					fmt.Println("-- ATRMultiplier:     ", result.IndicatorParams.BetterBands.ATRMultiplier)
					fmt.Println("-- BandMult2:         ", result.IndicatorParams.BetterBands.BandMult2)
					fmt.Println("-- BandMult3:         ", result.IndicatorParams.BetterBands.BandMult3)
					fmt.Println("-- MoveBoostParameter:", result.IndicatorParams.BetterBands.MoveBoostParameter)
					fmt.Println("-- MidlineTrendParam: ", result.IndicatorParams.BetterBands.MidlineTrendParam)
					fmt.Println("-- MidlineBoostParam: ", result.IndicatorParams.BetterBands.MidlineBoostParam)
				}
			}

			//Sort OptimiserResults by Pnl and number of trades
			sort.Slice(optimiserResults[symbol], func(i, j int) bool {
				return optimiserResults[symbol][i].Trades > optimiserResults[symbol][j].Trades
			})

			fmt.Println()
			fmt.Println("-----------------------------------")
			fmt.Println("          by TRADES ")
			fmt.Println()
			fmt.Println(data.SymbolNoUSDT(symbol))

			for i, result := range optimiserResults[symbol] {
				if i < 9 {
					fmt.Println()
					fmt.Println("-----------------------------------")
					fmt.Println(result.IterationName)
					fmt.Println("PnL:             ", fmt.Sprintf("%.2f%s", result.PnL, "%"))
					fmt.Println("Balance:         ", fmt.Sprintf("%.2f", result.Balance))
					fmt.Println("MaxDrawdown:     ", fmt.Sprintf("%.2f", result.MaxDrawdown))
					fmt.Println("WinRate:         ", fmt.Sprintf("%.2f%s", result.WinRate, "%"))
					fmt.Println("Trades:          ", result.Trades)
					fmt.Println()
					fmt.Println("TP, SL: ", result.StrategyParams.TP, result.StrategyParams.SL)
					fmt.Println("TrendMove:")
					fmt.Println("-- TrendPeriod:       ", result.IndicatorParams.TrendMove.TrendPeriod)
					fmt.Println("-- MovePeriod:        ", result.IndicatorParams.TrendMove.MovePeriod)
					fmt.Println("-- TrendSmoothing:       ", result.IndicatorParams.TrendMove.TrendSmoothing)
					fmt.Println("-- MoveSmoothing:        ", result.IndicatorParams.TrendMove.MoveSmoothing)
					fmt.Println("BetterBands:")
					fmt.Println("-- BandsPeriod:       ", result.IndicatorParams.BetterBands.BandsPeriod)
					fmt.Println("-- ATRMultiplier:     ", result.IndicatorParams.BetterBands.ATRMultiplier)
					fmt.Println("-- BandMult2:         ", result.IndicatorParams.BetterBands.BandMult2)
					fmt.Println("-- BandMult3:         ", result.IndicatorParams.BetterBands.BandMult3)
					fmt.Println("-- MoveBoostParameter:", result.IndicatorParams.BetterBands.MoveBoostParameter)
					fmt.Println("-- MidlineTrendParam: ", result.IndicatorParams.BetterBands.MidlineTrendParam)
					fmt.Println("-- MidlineBoostParam: ", result.IndicatorParams.BetterBands.MidlineBoostParam)
				}
			}

			fmt.Println()

		} else {
			fmt.Println()
			fmt.Println(symbol, "has no profitable combinations")
			fmt.Println()

		}

	}

	fmt.Println("Total time elapsed:", time.Since(timeTotal).Minutes())

	// -----------------------------------------------------------------------------
	// GENERATE LIST OF STRATEGY MODES

	//listOfStrategyModes := []system.StrategyMode{
	//	system.ModeHftModerate,
	//	system.ModeHftModerateWiderTpSl,
	//	system.ModeHftModerateTightTpSl,
	//	system.ModeHftModerateWiderTpTighterSl,
	//	system.ModeHftModerateWiderTpSlFast,
	//	system.ModeHftModerateWiderTpSlSlower,
	//	system.ModeHftWide,
	//	system.ModeHftNarrow,
	//	system.ModeHftCalm,
	//	system.ModeSlowCalm,
	//	system.ModeVolatileAggressive,
	//	system.ModeHftModerate_Narrow,
	//	system.ModeHftModerate_NarrowAndFaster,
	//	system.ModeHftModerate_NarrowAndSlower,
	//	system.ModeHftModerate_TwiceNarrow,
	//	system.ModeHftModerate_TwiceNarrowAndFaster,
	//	system.ModeHftModerate_TwiceNarrowAndSlower,
	//}

	//fmt.Println()
	//fmt.Println()
	//for symbol, results := range optimiserResults {
	//	//Sort OptimiserResults by Pnl and number of trades
	//	sort.Slice(results, func(i, j int) bool {
	//		return results[i].PnL > results[j].PnL
	//	})
	//
	//	if results[0].PnL > 0 && results[0].Trades > 10 {
	//		fmt.Println()
	//		fmt.Println("===================================")
	//		fmt.Println()
	//		fmt.Println(data.SymbolNoUSDT(symbol))
	//
	//		for i, result := range results {
	//			if i < 10 {
	//				fmt.Println()
	//				fmt.Println("-----------------------------------")
	//				fmt.Println(result.IterationName)
	//				fmt.Println("PnL:             ", fmt.Sprintf("%.2f%s", result.PnL, "%"))
	//				fmt.Println("Balance:         ", fmt.Sprintf("%.2f", result.Balance))
	//				fmt.Println("MaxDrawdown:     ", fmt.Sprintf("%.2f", result.MaxDrawdown))
	//				fmt.Println("WinRate:         ", fmt.Sprintf("%.2f%s", result.WinRate, "%"))
	//				fmt.Println("Trades:          ", result.Trades)
	//				fmt.Println()
	//				fmt.Println("TP, SL: ", result.StrategyParams.TP, result.StrategyParams.SL)
	//				fmt.Println("TrendMove:")
	//				fmt.Println("-- TrendPeriod:       ", result.IndicatorParams.TrendMove.TrendPeriod)
	//				fmt.Println("-- MovePeriod:        ", result.IndicatorParams.TrendMove.MovePeriod)
	//				fmt.Println("BetterBands:")
	//				fmt.Println("-- BandsPeriod:       ", result.IndicatorParams.BetterBands.BandsPeriod)
	//				fmt.Println("-- ATRMultiplier:     ", result.IndicatorParams.BetterBands.ATRMultiplier)
	//				fmt.Println("-- BandMult2:         ", result.IndicatorParams.BetterBands.BandMult2)
	//				fmt.Println("-- BandMult3:         ", result.IndicatorParams.BetterBands.BandMult3)
	//				fmt.Println("-- MoveBoostParameter:", result.IndicatorParams.BetterBands.MoveBoostParameter)
	//				fmt.Println("-- MidlineTrendParam: ", result.IndicatorParams.BetterBands.MidlineTrendParam)
	//				fmt.Println("-- MidlineBoostParam: ", result.IndicatorParams.BetterBands.MidlineBoostParam)
	//			}
	//		}
	//	}
	//}

}

// ----------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------

func backtestStrategyMode(candles *data.Candles, strategy system.Strategy) StrategyModeResults {
	// Data Persistence
	var trades oms.Trades

	startingAccountBalance := 100.0
	accountBalance := 100.0
	var maxDrawdown float64

	var openOrder oms.OrderLevel
	inventory := oms.SymbolInventory{
		Capacity:           1,
		ActiveTrades:       []*oms.ActiveTrade{},
		LastTimeOrderAdded: 0,
	}

	var strategyModeResults StrategyModeResults

	/* BACKTESTING PROCESS */
	// Loop over candles

	for _, candle := range *candles {
		//if accountBalance <= 0 {
		//	fmt.Println("\n\nLIQUIDATION: accountBalance is LESS than 0 at", len(trades), "trades")
		//	return StrategyModeResults{
		//		Balance:         0,
		//		PnL:             0,
		//		MaxDrawdown:     0,
		//		WinRate:         0,
		//		Trades:          0,
		//		StrategyParams:  strategy.StrategyParams,
		//		IndicatorParams: strategy.IndicatorParams,
		//	}
		//}
		// Send each candle to strategy
		// Strategy returns it's decision
		orderset := strategy.Evaluate(candle)

		// Decision is piped to omsSimulator
		// omsSimulator logs results into BacktestStats
		omsSimulator(&accountBalance, &inventory, &openOrder, &orderset, candle, &trades)

		if (accountBalance - startingAccountBalance) < maxDrawdown {
			maxDrawdown = accountBalance - startingAccountBalance
		}

	}

	strategyModeResults = calculateStrategyModeResults(&strategy, &trades, startingAccountBalance, accountBalance, maxDrawdown)

	return strategyModeResults
}

func calculateStrategyModeResults(strategy *system.Strategy, trades *oms.Trades, startingAccountBalance, accountBalance, maxDrawdown float64) StrategyModeResults {
	tradesCount := len(*trades)

	bbParams := *strategy.BetterBands
	tmParams := *strategy.TrendMove
	strategyParams := *strategy.StrategyParams

	indicatorParamsCopy := system.IndicatorParams{
		BetterBands: &bbParams,
		TrendMove:   &tmParams,
	}

	if tradesCount > 0 {
		sort.Slice(*trades, func(i, j int) bool {
			return (*trades)[i].ExitTime < (*trades)[j].ExitTime
		})

		var longsCount int
		var shortsCount int

		var longsTpCount int
		var longsSlCount int
		var shortsTpCount int
		var shortsSlCount int

		for _, t := range *trades {
			if t.Side == oms.SideLong {
				longsCount++
				// L O N G S
				// TP
				if t.ExitPrice > t.EntryPrice {
					longsTpCount++
				} else { // SL
					longsSlCount++
				}

			} else if t.Side == oms.SideShort {
				shortsCount++
				// S H O R T S
				// TP
				if t.ExitPrice < t.EntryPrice {
					shortsTpCount++
				} else { // SL
					shortsSlCount++
				}
			}
		}

		pnl := ta.PctChange(startingAccountBalance, accountBalance)
		totalWins := longsTpCount + shortsTpCount
		winRate := ta.PercentsXofY(float64(totalWins), float64(tradesCount))

		return StrategyModeResults{
			IterationName:   strategy.Name,
			Balance:         accountBalance,
			PnL:             pnl,
			MaxDrawdown:     maxDrawdown,
			WinRate:         winRate,
			Trades:          tradesCount,
			StrategyParams:  strategyParams,
			IndicatorParams: indicatorParamsCopy,
		}

	} else {
		return StrategyModeResults{
			IterationName:   strategy.Name,
			Balance:         0,
			PnL:             0,
			MaxDrawdown:     0,
			WinRate:         0,
			Trades:          0,
			StrategyParams:  strategyParams,
			IndicatorParams: indicatorParamsCopy,
		}
	}

}

// OptimiserResults Database of all StrategyModeResults for all symbols
// stored as map[symbol][]StrategyResults sorted by Pnl and Trades
type OptimiserResults map[string][]StrategyModeResults

// StrategyModeResults stats for single symbol and single strategy mode
type StrategyModeResults struct {
	IterationName string
	Balance       float64
	PnL           float64
	WinRate       float64
	MaxDrawdown   float64
	Trades        int
	system.StrategyParams
	system.IndicatorParams
}

func (smr *StrategyModeResults) ExportToJson() {

}

// ----------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------
// STRATEGY MODE DEV

// --- [0] Screener
var pvStrategy = map[string][]float64{
	"TP": {0.16, 0.2, 0.25, 0.3, 0.35},
	"SL": {0.45},
}

var pvBetterBands = map[string][]float64{
	"BandsPeriod":        {7, 10, 15, 20},
	"ATRMultiplier":      {5, 8, 10, 12, 14, 18},
	"BandMult2":          {1},
	"BandMult3":          {1},
	"MoveBoostParameter": {0.0008},
	"MidlineTrendParam":  {1.5},
	"MidlineBoostParam":  {1.8},
}

var pvTrendMove = map[string][]float64{
	"TrendPeriod":    {120},
	"MovePeriod":     {40},
	"TrendSmoothing": {21},
	"MoveSmoothing":  {14},
}

// --- [1] General Discovery
//var pvStrategy = map[string][]float64{
//	"TP": {0.2},
//	"SL": {0.45},
//}
//
//var pvBetterBands = map[string][]float64{
//	"BandsPeriod":        {7},
//	"ATRMultiplier":      {14},
//	"BandMult2":          {1},
//	"BandMult3":          {1},
//	"MoveBoostParameter": {0.0008},
//	"MidlineTrendParam":  {1.5},
//	"MidlineBoostParam":  {1.8},
//}
//
//var pvTrendMove = map[string][]float64{
//	"TrendPeriod":    {120},
//	"MovePeriod":     {40},
//	"TrendSmoothing": {21},
//	"MoveSmoothing":  {14},
//}

// --- [2.1] Periods and BB BandMult, MoveBoostParameter

//var pvStrategy = map[string][]float64{
//	"TP": {0.2},
//	"SL": {0.3},
//}
//
//var pvBetterBands = map[string][]float64{
//	"BandsPeriod":        {7},
//	"ATRMultiplier":      {10, 12, 14, 18},         // *
//	"BandMult2":          {1.5, 2, 2.5},            // *
//	"BandMult3":          {1.5, 2, 2.5},            // *
//	"MoveBoostParameter": {0.0006, 0.0007, 0.0009}, // *
//	"MidlineTrendParam":  {1.5},
//	"MidlineBoostParam":  {1.8},
//}
//
//var pvTrendMove = map[string][]float64{
//	"TrendPeriod":    {180}, // *
//	"MovePeriod":     {30},  // *
//	"TrendSmoothing": {21},
//	"MoveSmoothing":  {14},
//}

// --- [2.2] BB Details (MidlineTrendParam, MidlineBoostParam)

//var pvStrategy = map[string][]float64{
//	"TP": {0.24},
//	"SL": {0.35},
//}
//
//var pvBetterBands = map[string][]float64{
//	"BandsPeriod":        {5}, // *
//	"ATRMultiplier":      {10},
//	"BandMult2":          {2.5},
//	"BandMult3":          {2},
//	"MoveBoostParameter": {0.0009},
//	"MidlineTrendParam":  {1.5}, // *
//	"MidlineBoostParam":  {1.8}, // *
//}
//
//var pvTrendMove = map[string][]float64{
//	"TrendPeriod":    {80, 120, 180, 200},
//	"MovePeriod":     {20, 30, 40},
//	"TrendSmoothing": {18, 21, 29},
//	"MoveSmoothing":  {9, 14, 18},
//}

//var pvStrategy = map[string][]float64{
//	"TP": {0.24, 0.3, 0.35},
//	"SL": {0.24, 0.3, 0.35, 0.38, 0.4},
//}
//
//var pvBetterBands = map[string][]float64{
//	"BandsPeriod":        {5}, // *
//	"ATRMultiplier":      {10},
//	"BandMult2":          {2.5},
//	"BandMult3":          {2},
//	"MoveBoostParameter": {0.0009},
//	"MidlineTrendParam":  {1.5}, // *
//	"MidlineBoostParam":  {1.8}, // *
//}
//
//var pvTrendMove = map[string][]float64{
//	"TrendPeriod":    {200},
//	"MovePeriod":     {40},
//	"TrendSmoothing": {29},
//	"MoveSmoothing":  {14},
//}

// --- [3] TP SL fine-tuning
//var pvStrategy = map[string][]float64{
//	"TP": {0.18, 0.2, 0.24}, // *
//	"SL": {0.25, 0.3, 0.35}, // *
//}
//
//var pvBetterBands = map[string][]float64{
//	"BandsPeriod":        {5},
//	"ATRMultiplier":      {10, 12, 14},
//	"BandMult2":          {2.5},
//	"BandMult3":          {2},
//	"MoveBoostParameter": {0.0006, 0.0007, 0.0008, 0.001},
//	"MidlineTrendParam":  {1.5},
//	"MidlineBoostParam":  {1.8},
//}
//
//var pvTrendMove = map[string][]float64{
//	"TrendPeriod":    {180},
//	"TrendSmoothing": {21},
//	"MovePeriod":     {30},
//	"MoveSmoothing":  {14},
//}

// ------------------------------------

func calculateTotalNumberOfParameterVariations(pvStrategy, pvBetterBands, pvTrendMove map[string][]float64) int {
	var totalNum int = 1

	for _, xTrendPeriod := range pvTrendMove["TrendPeriod"] {
		for _, xMovePeriod := range pvTrendMove["MovePeriod"] {
			for _, xTrendSmoothing := range pvTrendMove["TrendSmoothing"] {
				for _, xMoveSmoothing := range pvTrendMove["MoveSmoothing"] {
					for _, xBandsPeriod := range pvBetterBands["BandsPeriod"] {
						for _, xATRMultiplier := range pvBetterBands["ATRMultiplier"] {
							for _, xBandMult2 := range pvBetterBands["BandMult2"] {
								for _, xBandMult3 := range pvBetterBands["BandMult3"] {
									for _, xMoveBoostParameter := range pvBetterBands["MoveBoostParameter"] {
										for _, xMidlineTrendParam := range pvBetterBands["MidlineTrendParam"] {
											for _, xMidlineBoostParam := range pvBetterBands["MidlineBoostParam"] {
												for _, xTP := range pvStrategy["TP"] {
													for _, xSL := range pvStrategy["SL"] {
														_ = xTrendPeriod + xMovePeriod + xTrendSmoothing + xMoveSmoothing + xBandsPeriod +
															xATRMultiplier + xBandMult2 + xBandMult3 + xMoveBoostParameter + xMidlineTrendParam +
															xMidlineBoostParam + xTP + xSL

														totalNum++
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return totalNum

}
