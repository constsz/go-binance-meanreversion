package main

import (
	"directionalMaker/data"
	"directionalMaker/system"
	"directionalMaker/ta"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func Chart() {
	cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println()

	symbol := "BNBUSDT"

	fmt.Println(symbol)

	//var (
	//	apiKey    = "CZYRAfunS6FVDQJH5cjZCQsA89SsiJrVqCN5G5DR8HGzyiVJXPw2k4ASChjR6Lgk"
	//	secretKey = "lGvXuV3iaYDo4LovFu03nOPCKguCFo83M2uZpVxRe9ANGcsPsyo0akZJOP7mVsKS"
	//)
	//client := binance.NewFuturesClient(apiKey, secretKey)
	//hoursWindow := 2

	sm := system.ModeVolatileAggressive

	//var candles data.Candles
	//candles.LoadHistoricalData(client, symbol, 90, hoursWindow)

	filePath := "D:\\programming\\Projects_Trading\\candles\\main\\candles_200\\CFXUSDT\\CFXUSDT-candles-2023-03-22.csv"
	candles := data.LoadCSVCandles(filePath)

	strategy := *system.NewStrategy(
		&system.StrategyParams{
			TF:               sm.Timeframe,
			Quantity:         sm.QtyUsd,
			TP:               sm.TP,
			SL:               sm.SL,
			LogicTP:          system.TP_Fixed,
			LogicSL:          system.SL_Fixed,
			OrderTimeSpacing: 1 * time.Second.Milliseconds(),
		},
		&system.IndicatorParams{
			BetterBands: &sm.BetterBandsParams,
			TrendMove:   &sm.TrendMoveParams,
		}, symbol, 0)

	strategy.Name = fmt.Sprintf("%s-%v", sm.Name, 0)

	// Initialize symbol indicators
	tm := ta.NewTrendMove(0, &sm.TrendMoveParams)
	atr := ta.NewATR(0, 60)
	bb := ta.NewBetterBands(0, &sm.BetterBandsParams)

	// Calculate indicators
	tm.InitialCalc(&candles)
	atr.InitialCalc(&candles)

	candles.CutUnused(tm.Window - bb.Window)
	bb.InitialCalc(&candles)
	candles.CutUnused(bb.Window)

	indicatorsParams := [][]string{
		bb.GetParams(), tm.GetParams(),
	}

	candles = candles[:len(candles)-1]

	_ = indicatorsParams

	candles.ExportToChart(2000, indicatorsParams)

	// Run optimiser's Backtest process and return results
	result := backtestStrategyMode(&candles, strategy)

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
