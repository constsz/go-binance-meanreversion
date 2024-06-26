package orderbook

import (
	"github.com/adshao/go-binance/v2/futures"
	"strconv"
)

type Orderbooks map[string]*Orderbook

type Orderbook struct {
	LastUpdateId int64
	Bids         []PriceLevel
	Asks         []PriceLevel
}

type PriceLevel struct {
	Price    float64
	Quantity float64
}

// ConvertOrderbookResponse converts orderbook response to my format
func ConvertOrderbookResponse(ob *futures.DepthResponse) *Orderbook {
	newOB := Orderbook{LastUpdateId: 0}

	for _, bid := range ob.Bids {
		p, _ := strconv.ParseFloat(bid.Price, 10)
		q, _ := strconv.ParseFloat(bid.Quantity, 10)
		level := PriceLevel{
			Price:    p,
			Quantity: q,
		}
		newOB.Bids = append(newOB.Bids, level)
	}
	for _, ask := range ob.Asks {
		p, _ := strconv.ParseFloat(ask.Price, 10)
		q, _ := strconv.ParseFloat(ask.Quantity, 10)
		level := PriceLevel{
			Price:    p,
			Quantity: q,
		}
		newOB.Asks = append(newOB.Asks, level)
	}

	return &newOB
}

// SearchAsk searches the slice for an index.
// If element not exists, it returns the index of an element BEFORE which
// new element should be inserted.
// If new element is bigger than all existing elements, returned
// index will be len(slice)+1
func SearchAsk(asks *[]PriceLevel, price float64) (found bool, index int) {
	start := 0
	end := len(*asks) - 1

	for {
		var midpoint int = start + (end-start)/2

		if price == (*asks)[midpoint].Price {
			return true, midpoint
		}

		if price > (*asks)[midpoint].Price {
			start = midpoint
		} else {
			end = midpoint
		}

		if end-start == 1 {
			//log.Println("CASE end-start")
			if price == (*asks)[start].Price {
				return true, start
			} else if price == (*asks)[end].Price {
				return true, end
			} else if price > (*asks)[end].Price {
				return false, end + 1
			} else if price < (*asks)[end-start].Price {
				return false, start
			} else if price > (*asks)[end-start].Price {
				return false, end
			}
		}
	}

}

func SearchBid(asks *[]PriceLevel, price float64) (found bool, index int) {
	start := 0
	end := len(*asks) - 1

	for {
		var midpoint int = start + (end-start)/2

		//fmt.Println(start, midpoint, end)

		if price == (*asks)[midpoint].Price {
			return true, midpoint
		}

		if price < (*asks)[midpoint].Price {
			start = midpoint
		} else {
			end = midpoint
		}

		if end-start == 1 {
			//log.Println("CASE end-start")
			if price == (*asks)[start].Price {
				return true, start
			} else if price == (*asks)[end].Price {
				return true, end
			} else if price < (*asks)[end].Price {
				return false, end + 1
			} else if price > (*asks)[end-start].Price {
				return false, start
			} else if price < (*asks)[end-start].Price {
				return false, end
			}
		}
	}

}
