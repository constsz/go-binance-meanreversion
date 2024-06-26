package ta

import (
	"directionalMaker/data"
	"log"
)

type RMA struct {
	Name string
	// Used for main calculations
	Window int
	Type   string
}

func NewRma(name string, period int) *RMA {
	return &RMA{
		Name:   name,
		Window: period,
		Type:   "RMA",
	}
}

func (rma *RMA) InitialCalc(candles *data.Candles) {
	var rmaPrev float64

	for i := range *candles {
		if i > rma.Window {
			// Prepare a sample-window of data for RMA
			sampleOfCandles := (*candles)[i-rma.Window+1 : i+1]

			// Include only required values
			prices := make([]float64, rma.Window)
			for j, tick := range sampleOfCandles {
				prices[j] = tick.Close
			}

			// Calculate RMA and record value to indicator's column of a candle
			rmaVal := rma.Calc(&prices, rmaPrev)
			(*candles)[i].TA[rma.Name] = append((*candles)[i].TA[rma.Name], rmaVal)

			rmaPrev = rmaVal
		}
	}
}

func CalcRMA(src *[]float64, prevRma float64) float64 {
	var rmaVal float64
	if len(*src) == 0 {
		log.Fatal("RMA: provided data is empty")
	}
	length := float64(len(*src))

	alpha := 1 / length

	for _, source := range *src {
		if prevRma == 0 {
			rmaVal = SMA(src)
		} else {
			rmaVal = alpha*source + (1-alpha)*prevRma
		}
	}

	return rmaVal
}

// Calc RMA = alpha * source + (1 - alpha) * prevRma, where alpha = 1 / length
func (rma *RMA) Calc(src *[]float64, prevRma float64) float64 {
	return CalcRMA(src, prevRma)
}

func (rma *RMA) GetName() string {
	return rma.Name
}
func (rma *RMA) GetType() string {
	return rma.Type
}
