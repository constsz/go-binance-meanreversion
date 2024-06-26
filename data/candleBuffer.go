package data

import "github.com/adshao/go-binance/v2/futures"

// CandleBuffer is a temporary container to collect all incoming trades from WebSocket
// Once the CandleBuffer is filled (by counter, which is the same as in normal
// candle aggregation methods), the slice of trades is passed to function
// candles.AggregateSingleCandle(trades []*futures.WsAggTradeEvent)
// Once the CandleBuffer is filled, it set's the counter to 0.
// So when next time it runs and sees the counter=0, it will reset
// the slice of trades to empty.
type CandleBuffer struct {
	Trades  []*futures.WsAggTradeEvent
	Counter int
}

func (cb *CandleBuffer) Fill(period int, t *futures.WsAggTradeEvent) bool {
	// Increment counter
	//numberOfTradesInTick := int(t.LastTradeID) - int(t.FirstTradeID) + 1

	// Reset new candle
	if cb.Counter == 0 {
		// Fresh start - reset trades slice to empty
		cb.Trades = nil
	}

	// Fill the candle with new trades
	if cb.Counter < period {
		// Add new trade
		cb.Trades = append(cb.Trades, t)
		cb.Counter += 1
	}

	// Check if this time the candle is filled
	if cb.Counter >= period {
		// Reset the candle counter
		cb.Counter = 0
		// Return filled=true
		return true
	}

	return false
}
