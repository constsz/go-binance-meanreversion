package system

import (
	"directionalMaker/data"
	"directionalMaker/oms"
	"directionalMaker/ta"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"math"
	"time"
)

type Processor struct {
	//receiverChannel <-chan *futures.WsAggTradeEvent
	workers map[string]*Worker
}

func NewProcessor(client *futures.Client, symbols []string, strategyList *StrategyList, chOrderset chan<- *oms.OrderSet, chLastPrice chan<- *data.LastPrice) *Processor {
	p := &Processor{
		//receiverChannel: make(<-chan *futures.WsAggTradeEvent),
		workers: make(map[string]*Worker),
	}
	for _, symbol := range symbols {
		w := &Worker{
			symbol:   symbol,
			strategy: (*strategyList)[symbol],

			chAggTrade:  make(chan *futures.WsAggTradeEvent),
			chOrderset:  chOrderset,
			chLastPrice: chLastPrice,
			client:      client,
		}
		p.workers[symbol] = w
	}

	p.Start()
	return p
}

func (p *Processor) Start() {
	for _, w := range p.workers {
		w.RunSymbol()
		time.Sleep(5 * time.Second)
	}
}

func (p *Processor) Listen(ch chan *futures.WsAggTradeEvent) {
	go func() {
		for {
			wsAggTrade := <-ch
			// Parse symbol from wsAggTrade
			aggTradeSymbol := wsAggTrade.Symbol
			// Route it to appropriate worker channel
			p.workers[aggTradeSymbol].chAggTrade <- wsAggTrade
		}
	}()
}

type Worker struct {
	symbol   string
	strategy Strategy

	chAggTrade  chan *futures.WsAggTradeEvent
	chOrderset  chan<- *oms.OrderSet
	chLastPrice chan<- *data.LastPrice
	client      *futures.Client
}

func (w *Worker) RunSymbol() {
	var candleIdCounter int
	// Initialize symbol dataframe
	var candles data.Candles
	// Initialize symbol indicators
	tm := ta.NewTrendMove(w.strategy.Id, w.strategy.TrendMove)
	atr := ta.NewATR(w.strategy.Id, 60)
	bb := ta.NewBetterBands(w.strategy.Id, w.strategy.BetterBands)

	// Preload historical candles
	candles.PreLoad(w.client, w.symbol, w.strategy.TF, tm.Window*2, &candleIdCounter)

	noPriceErrors := false

	// Calculate indicators
	tm.InitialCalc(&candles)
	atr.InitialCalc(&candles)

	candles.CutUnused(tm.Window - bb.Window)
	//fmt.Println("After CutUnused(tm.Window - bb.Window):", len(candles))

	bb.InitialCalc(&candles)

	//fmt.Println("After BB:", len(candles))

	candles.CutUnused(bb.Window)

	//fmt.Println("After CutUnused(bb.Window):", len(candles))

	fmt.Printf("\n%s\nPreloaded historical data.\n", w.symbol)
	fmt.Println("Listening...")
	fmt.Println()

	var candleBuffer data.CandleBuffer
	go func() {
		for {
			select {
			case msg := <-w.chAggTrade:
				// Aggregate incoming trades into slice and count counter
				filled := candleBuffer.Fill(w.strategy.TF, msg)

				if filled {
					// Pass candleBuffer to candles.AggregateSingleCandle
					candles.AggregateSingleCandle(candleBuffer)

					// Calculate indicators
					tm.CandleCalc(&candles)
					atr.CandleCalc(&candles)
					bb.CandleCalc(&candles)

					// Send last candle to strategy, which returns orderset
					orderset := w.strategy.Evaluate(candles[len(candles)-1])

					// check if no crazy errors in data
					if !noPriceErrors {
						priceErrorDifference := math.Abs(ta.PctChange(candles[len(candles)-1].Close, candles[len(candles)-2].Close))
						//fmt.Println("% CHNG:", priceErrorDifference)
						if priceErrorDifference < 0.005 {
							noPriceErrors = true
						}
					}

					// Channel orderset to OMS
					// Quick check if strategy is not doing anything crazy
					if noPriceErrors {
						w.chOrderset <- &orderset

					} else {
						//fmt.Println()
						//fmt.Println(w.symbol)
						//fmt.Println("Wild caught:")
						//fmt.Println("% dif:       ", ta.PctChange(orderset.OrderLevel.EntryPrice, candles[len(candles)-1].Close))
						//fmt.Println("Entry:       ", orderset.OrderLevel.EntryPrice)
						//fmt.Println("comparing close, if it changes (new, old):", candles[len(candles)-1].Close, candles[len(candles)-2].Close)
						//fmt.Println("time difference:", time.UnixMilli(candles[len(candles)-1].Time).Sub(time.UnixMilli(candles[len(candles)-2].Time)).Seconds())
					}

					//w.chLastPrice <- &data.LastPrice{
					//	Symbol: w.symbol,
					//	Close:  candles[len(candles)-1].Close,
					//}

					// Cut the candles slice when it's too big
					if len(candles) > tm.Window*4 {
						candles.CutToLength(tm.Window * 3)
					}
				}
			}
		}
	}()
}
