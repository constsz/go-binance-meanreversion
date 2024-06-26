package data

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"log"
	"math"
	"strconv"
	"time"
)

// MarketData is a global dataframe
// of all symbols tick candles: { SymbolName: []Candles }
type MarketData map[string]Candles

type Candles []Candle

type Candle struct {
	Id          int
	Time        int64
	Open        float64
	High        float64
	Low         float64
	Close       float64
	Quantity    float64
	VolumeDelta float64
	IsClosed    bool
	Counter     int
	TA          map[string][]float64
}

type AggTrade struct {
	AggTradeId   int64
	Price        float64
	Quantity     float64
	FirstTradeID int64
	LastTradeID  int64
	Time         int64
	BuyerMaker   bool
}

type LastPrice struct {
	Symbol string
	Close  float64
}

// CONVERTING EVENTS to internal format

// ConvertToAggTrade_WsAggTradeEvent converts WebSocket aggTradeEvent to internal format struct
func ConvertToAggTrade_WsAggTradeEvent(event *futures.WsAggTradeEvent) AggTrade {
	p, _ := strconv.ParseFloat(event.Price, 10)
	q, _ := strconv.ParseFloat(event.Quantity, 10)

	return AggTrade{
		AggTradeId:   event.AggregateTradeID,
		Price:        p,
		Quantity:     q,
		FirstTradeID: event.FirstTradeID,
		LastTradeID:  event.LastTradeID,
		Time:         event.Time,
		BuyerMaker:   event.Maker,
	}

}

// ConvertToAggTrade_AggTradeEvent converts REST API response aggTradeEvent to internal format struct
func ConvertToAggTrade_AggTradeEvent(event *futures.AggTrade) AggTrade {
	p, _ := strconv.ParseFloat(event.Price, 10)
	q, _ := strconv.ParseFloat(event.Quantity, 10)

	return AggTrade{
		AggTradeId:   event.AggTradeID,
		Price:        p,
		Quantity:     q,
		FirstTradeID: event.FirstTradeID,
		LastTradeID:  event.LastTradeID,
		Time:         event.Timestamp,
		BuyerMaker:   event.IsBuyerMaker,
	}

}

// DATA LOGIC METHODS

// AggregateHistoricalCandle receives single AggTrade struct and adds it to candles
func (candles *Candles) AggregateHistoricalCandle(aggTrade AggTrade, period int, candleIdCounter *int) {
	// Because we use AggTrades, when for example you have Candle period == 100,
	// and you have existing Candle at 99, when you receive new AggTrade with 3 trades in it,
	// all 3 trades will be added to open candle, so it will close with 102 trades count.
	// Not critical at the moment, let's ignore for now.
	numberOfTradesInTick := int(aggTrade.LastTradeID) - int(aggTrade.FirstTradeID) + 1

	// If {candles} is empty, create the first candle
	if len(*candles) == 0 {
		*candles = append(*candles, Candle{
			TA: make(map[string][]float64),
		})
	}
	last := len(*candles) - 1

	// Close and Quantity is updated on every incoming trade update
	(*candles)[last].Close = aggTrade.Price
	(*candles)[last].Quantity = (*candles)[last].Quantity + aggTrade.Quantity
	// Calculate Volume Delta
	if aggTrade.BuyerMaker {
		(*candles)[last].VolumeDelta = (*candles)[last].VolumeDelta - aggTrade.Quantity
	} else {
		(*candles)[last].VolumeDelta = (*candles)[last].VolumeDelta + aggTrade.Quantity
	}

	// Calculations for different Candle state by candleTickCounter number
	// if New Candle
	if (*candles)[last].Counter == 0 {
		// Set the candle open time
		(*candles)[last].Time = aggTrade.Time

		// Prices
		(*candles)[last].Open = aggTrade.Price
		(*candles)[last].High = aggTrade.Price
		(*candles)[last].Low = aggTrade.Price

		(*candles)[last].Counter += numberOfTradesInTick
	} else if (*candles)[last].Counter > 0 && (*candles)[last].Counter < period {
		// if we continue existing Candle
		// Candle Update
		if (*candles)[last].Close >= (*candles)[last].High {
			(*candles)[last].High = (*candles)[last].Close
		} else if (*candles)[last].Close <= (*candles)[last].Low {
			(*candles)[last].Low = (*candles)[last].Close
		}

		(*candles)[last].Counter += numberOfTradesInTick
	}

	if (*candles)[last].Counter >= period {
		// if closing Candle
		(*candles)[last].IsClosed = true

		if (*candles)[last].Close >= (*candles)[last].High {
			(*candles)[last].High = (*candles)[last].Close
		} else if (*candles)[last].Close <= (*candles)[last].Low {
			(*candles)[last].Low = (*candles)[last].Close
		}

		*candleIdCounter++
		(*candles)[last].Id = *candleIdCounter

		// This candle is closed, for the next iteration we create new candle.
		*candles = append(*candles, Candle{
			TA: make(map[string][]float64),
		})
	}
}

func (candles *Candles) AggregateSingleCandle(cb CandleBuffer) {
	b := cb.Trades

	// Do all work in one loop
	time := b[0].Time
	counter := cb.Counter
	isClosed := true

	open := ConvertToAggTrade_WsAggTradeEvent(b[0]).Price
	cClose := ConvertToAggTrade_WsAggTradeEvent(b[len(b)-1]).Price

	high := 0.0
	low := math.MaxFloat64

	var quantity float64
	var volumeDelta float64

	for _, bt := range b {
		t := ConvertToAggTrade_WsAggTradeEvent(bt)
		if t.Price > high {
			high = t.Price
		}
		if t.Price < low {
			low = t.Price
		}
		quantity += t.Quantity

		// VolumeDelta
		if t.BuyerMaker {
			volumeDelta -= t.Quantity
		} else {
			volumeDelta += t.Quantity
		}
	}

	*candles = append(*candles, Candle{
		Time:        time,
		Open:        open,
		High:        high,
		Low:         low,
		Close:       cClose,
		Quantity:    quantity,
		VolumeDelta: volumeDelta,
		IsClosed:    isClosed,
		Counter:     counter,
		TA:          make(map[string][]float64, 3),
	})

}

func (candles *Candles) LoadHistoricalData(client *futures.Client, symbol string, candlePeriod int, periodHours int) {
	var timestampReached bool
	var firstTradeIdInSlice int64
	var tradesAll []*futures.AggTrade

	for !timestampReached {
		var fromId int64
		if firstTradeIdInSlice > 0 {
			fromId = firstTradeIdInSlice - 1000
		}

		// 1. Load first 1000 of trades
		trades, err := futuresLoadAggTrades(client, symbol, fromId)
		//fmt.Println(symbol, " | Binance: downloaded", len(trades), "trades from ", time.UnixMilli(trades[0].Timestamp).Local(), time.UnixMilli(trades[len(trades)-1].Timestamp).Local())
		time.Sleep(250 * time.Millisecond)
		if err != nil {
			log.Println("PreLoad: Error during Binance API request:\n", err)
			return
		}

		// 2. Combine with rest of the Trades (prepend to start of slice)
		tradesAll = append(trades, tradesAll...)

		firstTradeIdInSlice = trades[0].AggTradeID

		if trades[0].Timestamp <= time.Now().UnixMilli()-(int64(periodHours)*time.Hour.Milliseconds()) {
			timestampReached = true
		}
	}

	// 3. Aggregate candles
	for _, t := range tradesAll {
		id := 0
		aggTrade := ConvertToAggTrade_AggTradeEvent(t)
		candles.AggregateHistoricalCandle(aggTrade, candlePeriod, &id)
	}

}

func (candles *Candles) PreLoad(client *futures.Client, symbol string, period int, minNumOfCandles int, candleIdCounter *int) {
	iteration := 1
	var loadingCandleTickCounter int
	var loadedCandles int

	minNumOfCandles += 50

	var firstTradeIdInSlice int64
	//var lastTradeIdInSlice int64
	var tradesAll []*futures.AggTrade

	//fmt.Println("\nPreloading historical candles ...")

	// Start loop, that will load up trades until the counter reaches numberOfCandles target
	for loadedCandles < minNumOfCandles {
		if iteration > 1 {
			time.Sleep(50 * time.Millisecond)
		}
		//fmt.Println("Initeration ", iteration)

		var fromId int64
		if firstTradeIdInSlice > 0 {
			fromId = firstTradeIdInSlice - 1000
		}

		// 1. Load first 1000 of trades
		trades, err := futuresLoadAggTrades(client, symbol, fromId)
		if err != nil {
			log.Println("PreLoad: Error during Binance API request:\n", err)
			return
		}

		// 2. Combine with rest of the Trades (prepend to start of slice)
		tradesAll = append(trades, tradesAll...)

		// 3. Increment Counters
		for _, t := range trades {
			numberOfTradesInTick := int(t.LastTradeID) - int(t.FirstTradeID) + 1

			if loadingCandleTickCounter < period {
				loadingCandleTickCounter += numberOfTradesInTick
			} else if loadingCandleTickCounter >= period {
				loadedCandles++
				// Reset the candle candle and candleTickCounter
				loadingCandleTickCounter = 0
			}

		}

		firstTradeIdInSlice = trades[0].AggTradeID
		//lastTradeIdInSlice = trades[len(trades)-1].AggTradeID
		iteration++

		//fmt.Println("loaded candles count:", loadedCandles)
		//fmt.Println("firstTradeIdInSlice:", firstTradeIdInSlice)
		//fmt.Println("lastTradeIdInSlice: ", lastTradeIdInSlice)

	}

	// 4. Aggregate candles
	for _, t := range tradesAll {
		aggTrade := ConvertToAggTrade_AggTradeEvent(t)
		candles.AggregateHistoricalCandle(aggTrade, period, candleIdCounter)
	}

	//fmt.Println("\nCompare counts: trades vs candles")
	//fmt.Println("len(*candles)                |", len(*candles))
	//fmt.Println("len(tradesAll)               |", len(tradesAll))

	var tradesCountNew int
	for _, t := range tradesAll {
		tradesCountNew = tradesCountNew + (int(t.LastTradeID) - int(t.FirstTradeID) + 1)
	}
	//fmt.Println("NUM of candles by tradesAll  |", tradesCountNew/period)
}

func futuresLoadAggTrades(client *futures.Client, symbol string, fromId int64) ([]*futures.AggTrade, error) {
	var trades []*futures.AggTrade
	var err error

	if fromId == 0 {
		trades, err = client.NewAggTradesService().
			Symbol(symbol).Limit(1000).Do(context.Background())
		if err != nil {
			log.Println("PreLoad: Error during Binance API request:\n", err)
		}
	} else {
		trades, err = client.NewAggTradesService().
			Symbol(symbol).FromID(fromId).Limit(1000).Do(context.Background())
		if err != nil {
			log.Println("PreLoad: Error during Binance API request:\n", err)
		}
	}

	return trades, err
}

func (candles *Candles) CutUnused(indicatorsMaxPeriod int) {
	*candles = append(Candles{}, (*candles)[indicatorsMaxPeriod:]...)
}

func (candles *Candles) CutToLength(length int) {
	last := len(*candles) - 1
	//a = append([]int(nil), a[3:]…)
	*candles = append(Candles{}, (*candles)[last-length:]...)
	//*candles = (*candles)[last-indicatorsMaxPeriod:]
}

func (candles *Candle) Print() {
	fmt.Println()
	fmt.Println("Time:       ", candles.Time)
	fmt.Println("Open:       ", candles.Open)
	fmt.Println("High:       ", candles.High)
	fmt.Println("Low:        ", candles.Low)
	fmt.Println("Close:      ", candles.Close)
	fmt.Println("Quantity:   ", candles.Quantity)
	fmt.Println("VolumeDelta:", candles.VolumeDelta)
	fmt.Println("Counter:    ", candles.Counter)
}
