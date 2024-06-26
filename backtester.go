package main

import (
	"directionalMaker/data"
	"directionalMaker/oms"
	"directionalMaker/system"
	"directionalMaker/ta"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

/*

For backtesting we need a list of strategies (each is a strategy instance with it's internal parameters).

The loop goes like this:
	- generate list of strategy params and indicator params -> generate a list of strategy instances
		- to generate list of params create a function, that has a list variations for each parameter
		- create multidimensional loop over all parameter lists to combine to parameter object and add it into global list
	- create instance of global backtesting results

	- for each TF
		- for each Symbol
			- load csv candles into one dataframe
			- process <all> indicators into this single dataframe using indicator parameters
			  in each strategy in the list. Name of indicator is a meta-name that serves as id.

			- Backtest this tf-symbol by looping over list of strategies using same single symbol-dataframe
			- Write backtesting results to backtesting global results

*/

func Backtesting() {
	start := time.Now()

	// Settings
	//dataFolder := "../candles/sample"
	dataFolder := "../candles/main"
	//symbols := []string{"APT"}
	symbols := data.GetListOfSymbols("../candles/main/candles_20")
	timeframes := []string{"20"}

	//whatToChart := map[string]string{
	//	"symbol":   "OP",
	//	"tf":       "100",
	//	"fileDate": "2023-01",
	//}

	bgr := NewBacktestGlobalResults(symbols)

	symbolBudget := SymbolBudget{
		Budget:   1000,
		Quantity: 500,
	}

	// Generate strategyList
	// TEMP: using just single strategy
	strategyList := []*system.Strategy{
		system.NewStrategy(
			system.StrategyDefaultParams(20, symbolBudget.Quantity),
			&system.IndicatorParams{
				BetterBands: ta.BetterBandsDefaultParams(),
				TrendMove:   ta.TrendMoveDefaultParams(),
			}, "BACKTEST", 1),
	}

	maxGoroutines := 6 //6
	guard := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	for _, symbol := range symbols {
		for _, tf := range timeframes {

			strategy := strategyList[0]
			strategy.Quantity = symbolBudget.Quantity

			symbolCsvFolder := filepath.Join(dataFolder, "candles_"+tf, data.SymbolUSDT(symbol))
			// Get a list of csv files with candles for each month available
			files := monthCsvCandles(symbolCsvFolder)

			// for each month csv file start backtesting
			for _, monthFile := range files {
				wg.Add(1)
				// go: each tf+symbol+month+strategyInstance -> is a separate goroutine with maxGoroutines limit
				go func(symbol string, tf string, strategy *system.Strategy, symbolCsvFolder string, monthFile os.DirEntry, wg *sync.WaitGroup) {
					guard <- struct{}{} // would block if guard channel is already filled
					fmt.Println("Started Goroutine for", symbol, tf, strategy.Id, data.ParseDateFromCsvFilename(symbol, monthFile.Name()))

					filePath := filepath.Join(symbolCsvFolder, monthFile.Name())
					fileDate := data.ParseDateFromCsvFilename(symbol, monthFile.Name())

					// Load CSV candles
					candles := data.LoadCSVCandles(filePath)

					// Calculate all indicators
					// ...
					// TEMP: testing indicators
					tm := ta.NewTrendMove(strategy.Id, strategy.TrendMove)
					tm.InitialCalc(&candles)

					atr := ta.NewATR(strategy.Id, 60)
					atr.InitialCalc(&candles)

					bb := ta.NewBetterBands(strategy.Id, strategy.BetterBands)
					candles.CutUnused(tm.Window - bb.Window)
					bb.InitialCalc(&candles)
					candles.CutUnused(bb.Window)

					backtest(&candles, strategy, bgr, BacktestInfo{
						Symbol:       symbol,
						SymbolBudget: symbolBudget,
						TF:           tf,
						CandlesDate:  fileDate,
					})

					//if symbol == whatToChart["symbol"] && tf == whatToChart["tf"] && fileDate == whatToChart["fileDate"] {
					//	// Export to chart
					//	candles.ExportToChart(20, nil)
					//}

					<-guard
					wg.Done()
				}(symbol, tf, strategy, symbolCsvFolder, monthFile, &wg)

			}

			//bgr.PrintSymbolStats(symbol)
		}
	}

	wg.Wait()
	elapsed := time.Since(start)

	finStat := bgr.CreateBacktestFinalStats()
	finStat.Print()

	fmt.Println()
	fmt.Printf("\nExecution time: %f min.\n", elapsed.Minutes())

}

// -------------------------------------------------------------
// GENERATE STRATEGY LIST
// These are hard-coded use case scripts, not universal methods.
func generateStrategyList() (listOfStrategies []*system.Strategy) {
	strategyIdCounter := 0
	indicatorParamsList := generateIndicatorParameterVariations()
	strategyParamsList := generateStrategyParameterVariations()

	for _, indParam := range indicatorParamsList {
		for _, stratParam := range strategyParamsList {
			strategy := system.NewStrategy(&stratParam, &indParam, "BACKTEST", strategyIdCounter)
			listOfStrategies = append(listOfStrategies, strategy)
			strategyIdCounter++
		}
	}

	return
}

func generateIndicatorParameterVariations() []system.IndicatorParams {
	// TrendMove Parameters
	// trendPeriodParams := []int{100,120,160,180,200}
	// ...

	// BetterBands Parameters
	// ...

	// Calculate totalLength by multiplying lengths of all parameter slices
	// indicatorParamsList := make([]system.IndicatorParams, totalLength)

	// Loop over each parameter list to combine the IndicatorParams object and append it to list

	return []system.IndicatorParams{}
}
func generateStrategyParameterVariations() []system.StrategyParams {
	return []system.StrategyParams{}
}

// -------------------------------------------------------------
// BACKTESTING RESULTS

func NewBacktestGlobalResults(symbols []string) *BacktestGlobalResults {
	backtestGlobalResults := BacktestGlobalResults{
		Symbol: make(map[string][]*BacktestResult, len(symbols)),
	}

	for _, symbol := range symbols {
		var backtestSymbolResults []*BacktestResult
		backtestGlobalResults.Symbol[symbol] = backtestSymbolResults
	}

	return &backtestGlobalResults
}

func (bgr *BacktestGlobalResults) InsertSymbolResult(symbol string, result *BacktestResult) {
	bgr.mu.Lock()
	defer bgr.mu.Unlock()
	bgr.Symbol[symbol] = append(bgr.Symbol[symbol], result)
}

func (bgr *BacktestGlobalResults) PrintSymbolStats(symbol string) {
	for i, s := range bgr.Symbol[symbol] {
		fmt.Println("\n_____________", symbol, "________________")
		fmt.Println(bgr.Symbol[symbol][i].CandlesDate)
		fmt.Println()
		fmt.Println("PnL           |", s.Stats.PnL)
		fmt.Println("AvgDaily      |", s.Stats.AvgDaily)
		fmt.Println("WinRate       |", s.Stats.WinRate)
		fmt.Println("MaxDrawdown   |", s.Stats.MaxDrawdown)
		fmt.Println()
		fmt.Println("Trades        |", s.Stats.Trades)
		fmt.Println("TradesPerDay  |", s.Stats.TradesPerDay)
		fmt.Println()
		fmt.Println("LongsPct      |", s.Stats.LongsPct)
		fmt.Println()
		fmt.Println("LongsPnL      |", s.Stats.LongsPnL)
		fmt.Println("LongsWR       |", s.Stats.LongsWR)
		fmt.Println("ShortsPnL     |", s.Stats.ShortsPnL)
		fmt.Println("ShortsWR      |", s.Stats.ShortsWR)
	}
}

func (bgr *BacktestGlobalResults) CreateBacktestFinalStats() *BacktestFinalStats {
	var finStat BacktestFinalStats

	for symbolName, symbolResults := range bgr.Symbol {
		var pnlAvgByMonth float64
		var pnlSum float64
		pnlWorstCase := 999999999999.0
		pnlBestCase := 0.0
		var tradesTotal int
		var tradesAvgByMonth int
		var balancesByMonth []float64

		for _, result := range symbolResults {
			balancesByMonth = append(balancesByMonth, math.Round(result.Stats.Balance)-1000)

			if result.Stats.PnL < pnlWorstCase {
				pnlWorstCase = result.Stats.PnL
			}
			if result.Stats.PnL > pnlBestCase {
				pnlBestCase = result.Stats.PnL
			}
			pnlSum += result.Stats.PnL
			tradesTotal += result.Stats.Trades
		}

		pnlAvgByMonth = pnlSum / float64(len(symbolResults))
		tradesAvgByMonth = tradesTotal / len(symbolResults)

		finStat = append(finStat, BacktestFinalSymbolStats{
			Symbol:           symbolName,
			BalanceByMonths:  balancesByMonth,
			PnlAvg:           pnlAvgByMonth,
			PnlWorstCase:     pnlWorstCase,
			PnlBestCase:      pnlBestCase,
			TradesTotal:      tradesTotal,
			TradesAvgByMonth: tradesAvgByMonth,
		})

	}

	sort.Slice(finStat, func(i, j int) bool {
		return finStat[i].PnlAvg > finStat[j].PnlAvg
	})

	return &finStat
}

func (finStat *BacktestFinalStats) Print() {
	fmt.Println()
	fmt.Println("-----------------------------------------")
	fmt.Println("TOTAL STATISTICS PER SYMBOL")
	fmt.Println("TFs and Strategies merged. Use it to see what symbols perform best.")

	for _, stats := range *finStat {
		fmt.Println()
		fmt.Println("----------------------------")
		fmt.Println(stats.Symbol)
		fmt.Println()
		fmt.Println("BalanceByMonths:  ", stats.BalanceByMonths)
		fmt.Println("PnlAvg:           ", fmt.Sprintf("%.2f", stats.PnlAvg))
		fmt.Println("PnlWorstCase:     ", fmt.Sprintf("%.2f", stats.PnlWorstCase))
		fmt.Println("PnlBestCase:      ", fmt.Sprintf("%.2f", stats.PnlBestCase))
		fmt.Println("TradesTotal:      ", stats.TradesTotal)
		fmt.Println("TradesAvgByMonth: ", stats.TradesAvgByMonth)
	}

}

type BacktestFinalStats []BacktestFinalSymbolStats

type BacktestFinalSymbolStats struct {
	Symbol           string
	BalanceByMonths  []float64
	PnlAvg           float64
	PnlWorstCase     float64
	PnlBestCase      float64
	TradesTotal      int
	TradesAvgByMonth int
}

type BacktestGlobalResults struct {
	mu     sync.Mutex
	Symbol map[string][]*BacktestResult
}

type BacktestResult struct {
	Symbol      string
	TF          string
	Strategy    string
	CandlesDate string
	Stats       BacktestStats
}

// -----------------------------

type BacktestStats struct {
	Balance      float64
	PnL          float64
	AvgDaily     float64
	WinRate      float64
	MaxDrawdown  float64
	Trades       int
	TradesPerDay int
	LongsPct     float64
	LongsPnL     float64
	LongsWR      float64
	ShortsPnL    float64
	ShortsWR     float64
}

type BacktestInfo struct {
	Symbol       string
	SymbolBudget SymbolBudget
	TF           string
	CandlesDate  string
}

func calculateBacktestStats(trades *oms.Trades, accountBalance float64) BacktestStats {
	fee_limit := 0.02
	fee_market := 0.04

	tradesCount := len(*trades)

	if tradesCount > 0 {
		sort.Slice(*trades, func(i, j int) bool {
			return (*trades)[i].ExitTime < (*trades)[j].ExitTime
		})

		day := 24 * time.Hour.Milliseconds()
		numOfDaysTraded := ((*trades)[tradesCount-1].ExitTime - (*trades)[0].ExitTime) / day
		if numOfDaysTraded == 0 {
			numOfDaysTraded = 1
		}
		tradesPerDay := int(int64(tradesCount) / numOfDaysTraded)

		var longsCount int
		var shortsCount int
		for _, t := range *trades {
			if t.Side == oms.SideLong {
				longsCount++
			} else if t.Side == oms.SideShort {
				shortsCount++
			}
		}

		longsPct := ta.PercentsXofY(float64(longsCount), float64(tradesCount))

		var longsTpCount int
		var longsSlCount int
		var shortsTpCount int
		var shortsSlCount int

		var longsPnL float64
		var shortsPnL float64

		var maxDrawdown float64

		for _, t := range *trades {
			if t.Side == oms.SideLong {
				// L O N G S
				// TP
				if t.ExitPrice > t.EntryPrice {
					longsTpCount++
					longsPnL += ta.PctChange(t.EntryPrice, t.ExitPrice) - fee_limit

				} else { // SL
					longsSlCount++
					loss := ta.PctChange(t.EntryPrice, t.ExitPrice) - fee_market
					longsPnL += loss
					if loss < maxDrawdown {
						maxDrawdown = loss
					}
				}

			} else if t.Side == oms.SideShort {
				// S H O R T S
				// TP
				if t.ExitPrice < t.EntryPrice {
					shortsTpCount++
					shortsPnL += ta.PctChange(t.ExitPrice, t.EntryPrice) - fee_limit
				} else { // SL
					shortsSlCount++
					loss := ta.PctChange(t.ExitPrice, t.EntryPrice) - fee_market
					shortsPnL += loss
					if loss < maxDrawdown {
						maxDrawdown = loss
					}
				}
			}
		}

		pnl := longsPnL + shortsPnL
		avgDaily := pnl / float64(numOfDaysTraded)

		totalWins := longsTpCount + shortsTpCount
		winRate := ta.PercentsXofY(float64(totalWins), float64(tradesCount))
		longsWR := ta.PercentsXofY(float64(longsTpCount), float64(longsCount))
		shortsWR := ta.PercentsXofY(float64(shortsTpCount), float64(shortsCount))

		return BacktestStats{
			Balance:      accountBalance,
			PnL:          pnl,
			AvgDaily:     avgDaily,
			WinRate:      winRate,
			MaxDrawdown:  maxDrawdown,
			Trades:       tradesCount,
			TradesPerDay: tradesPerDay,
			LongsPct:     longsPct,
			LongsPnL:     longsPnL,
			LongsWR:      longsWR,
			ShortsPnL:    shortsPnL,
			ShortsWR:     shortsWR,
		}
	} else {
		return BacktestStats{}
	}
}

func backtest(candles *data.Candles, strategy *system.Strategy, bgr *BacktestGlobalResults, backtestInfo BacktestInfo) {
	// Data Persistence
	var trades oms.Trades
	accountBalance := backtestInfo.SymbolBudget.Budget
	inventory := oms.SymbolInventory{
		Capacity:           10,
		ActiveTrades:       []*oms.ActiveTrade{},
		LastTimeOrderAdded: 0,
	}

	var openOrder oms.OrderLevel

	/* BACKTESTING PROCESS */
	// Loop over candles
	for _, candle := range *candles {
		if accountBalance <= 0 {
			fmt.Println("accountBalance is LESS than 0")
			return
		}
		// Send each candle to strategy
		// Strategy returns it's decision
		orderset := strategy.Evaluate(candle)
		// Decision is piped to omsSimulator
		// omsSimulator logs results into BacktestStats
		omsSimulator(&accountBalance, &inventory, &openOrder, &orderset, candle, &trades)
	}

	stats := calculateBacktestStats(&trades, accountBalance)
	// Record Backtest Result
	backtestResult := &BacktestResult{
		Symbol:      backtestInfo.Symbol,
		TF:          backtestInfo.TF,
		Strategy:    strategy.Name,
		CandlesDate: backtestInfo.CandlesDate,
		Stats:       stats,
	}

	bgr.InsertSymbolResult(backtestInfo.Symbol, backtestResult)

}

type SymbolBudget struct {
	Budget   float64
	Quantity float64
}

func omsSimulator(accountBalance *float64, inventory *oms.SymbolInventory, openOrder *oms.OrderLevel, orderset *oms.OrderSet, c data.Candle, trades *oms.Trades) {
	fee_limit := 0.02
	fee_market := 0.04
	marketStopSlippage := 0.14

	qty := openOrder.Quantity

	// Check SymbolInventory: if new candle price crossed TP or SL triggers
	// 		- If order exited by SL or TP:
	//				-- append to list of trades
	//				-- delete from inventory
	if len(inventory.ActiveTrades) > 0 {
		//inventoryLoop:
		for i, t := range inventory.ActiveTrades {
			if *accountBalance <= 0 {
				return
			}

			// StopLoss first
			if t.Side == oms.SideLong && c.Low < t.SL {
				ft := oms.Trade{
					Side:       oms.SideLong,
					EntryTime:  t.EntryTime,
					ExitTime:   c.Time,
					EntryPrice: t.EntryPrice,
					ExitPrice:  t.SL,
				}

				loss := (qty * ta.Pct(ta.PctChange(ft.ExitPrice, ft.EntryPrice)+fee_market+marketStopSlippage)) - qty
				*accountBalance -= loss

				*trades = append(*trades, ft)
				inventory.RemoveActiveOrder(i)
				//goto inventoryLoop
			} else if t.Side == oms.SideShort && c.High > t.SL {
				ft := oms.Trade{
					Side:       oms.SideShort,
					EntryTime:  t.EntryTime,
					ExitTime:   c.Time,
					EntryPrice: t.EntryPrice,
					ExitPrice:  t.SL,
				}

				loss := (qty * ta.Pct(ta.PctChange(ft.EntryPrice, ft.ExitPrice)+fee_market+marketStopSlippage)) - qty
				*accountBalance -= loss

				*trades = append(*trades, ft)
				inventory.RemoveActiveOrder(i)
				//goto inventoryLoop

				// TakeProfit
			} else if t.Side == oms.SideLong && c.High > t.TP {
				ft := oms.Trade{
					Side:       oms.SideLong,
					EntryTime:  t.EntryTime,
					ExitTime:   c.Time,
					EntryPrice: t.EntryPrice,
					ExitPrice:  t.TP,
				}

				profit := (qty * ta.Pct(ta.PctChange(ft.EntryPrice, ft.ExitPrice)-fee_limit)) - qty
				*accountBalance += profit

				*trades = append(*trades, ft)
				inventory.RemoveActiveOrder(i)
				//goto inventoryLoop
			} else if t.Side == oms.SideShort && c.Low < t.TP {
				ft := oms.Trade{
					Side:       oms.SideShort,
					EntryTime:  t.EntryTime,
					ExitTime:   c.Time,
					EntryPrice: t.EntryPrice,
					ExitPrice:  t.TP,
				}

				profit := (qty * ta.Pct(ta.PctChange(ft.ExitPrice, ft.EntryPrice)-fee_limit)) - qty
				*accountBalance += profit

				*trades = append(*trades, ft)
				inventory.RemoveActiveOrder(i)
				//goto inventoryLoop
			}
		}
	}

	if *accountBalance <= 0 {
		return
	}

	// Check openOrder: if new candle price crossed the current open order.
	// 		- If order was executed: add order to SymbolInventory and set openOrder to empty.
	if len(inventory.ActiveTrades) == 0 {
		if openOrder.Side == oms.SideLong && c.Low < openOrder.EntryPrice {
			newActiveTrade := oms.ActiveTrade{
				Side:       oms.SideLong,
				EntryTime:  c.Time,
				EntryPrice: openOrder.EntryPrice,
				TP:         openOrder.EntryPrice * ta.Pct(openOrder.TP),
				SL:         openOrder.EntryPrice / ta.Pct(openOrder.SL),
			}
			inventory.ActiveTrades = append(inventory.ActiveTrades, &newActiveTrade)
		} else if openOrder.Side == oms.SideShort && c.High > openOrder.EntryPrice {
			newActiveTrade := oms.ActiveTrade{
				Side:       oms.SideShort,
				EntryTime:  c.Time,
				EntryPrice: openOrder.EntryPrice,
				TP:         openOrder.EntryPrice / ta.Pct(openOrder.TP),
				SL:         openOrder.EntryPrice * ta.Pct(openOrder.SL),
			}
			inventory.ActiveTrades = append(inventory.ActiveTrades, &newActiveTrade)
		}
	}

	// Check if orderset.InstantAction requires any action.

	// Update openOrder with new order from order set.
	if len(inventory.ActiveTrades) == 0 && c.Time > inventory.LastTimeOrderAdded+orderset.OrderTimeSpacing {
		*openOrder = orderset.OrderLevel
	} else {
		openOrder.Side = oms.SideNone
	}

}

// -------------------------------------------------------------
// FILES
func monthCsvCandles(symbolCsvFolder string) []os.DirEntry {
	// Check if symbol csv folder exists
	_, err := os.Stat(symbolCsvFolder)
	if err != nil {
		log.Fatal("Folder not exist")
	}

	// Get a list of available months of candles for a symbol in a given timeframe
	files, err := os.ReadDir(symbolCsvFolder)
	if err != nil {
		log.Fatal(err)
	}

	return files
}
