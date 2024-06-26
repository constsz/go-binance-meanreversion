package main

import (
	"flag"
	"fmt"
	"runtime/debug"
)

func main() {
	debug.SetMemoryLimit(3e+9)
	flagOptimizer := flag.Bool("o", false, "Run Optimizer mode if true")
	flagChart := flag.Bool("c", false, "Run Optimizer mode if true")
	flagDebug := flag.Bool("d", false, "Runs section with custom functions")
	flag.Parse()

	if *flagOptimizer {
		fmt.Println("Starting Optimizer")
		Optimizer()
	} else if *flagDebug {
		fmt.Println("Debug Section is empty.")
		debugSection()
	} else if *flagChart {
		fmt.Println("Starting Chart")
		Chart()
	} else {
		fmt.Println("Starting LiveTrader")
		LiveTrader()
	}

}

func debugSection() {
	//Backtesting()
	//expUserData()
	//expExchangeData()
	//expTestingConvertFloatToBinancePrice()
	//expCreateOrders()
	//expGetSymbolOrders()
}
