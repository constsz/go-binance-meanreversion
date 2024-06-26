package ta

import (
	"directionalMaker/data"
	"log"
)

type EMA struct {
	Name string
	// Used for main calculations
	Window int
	Type   string
}

func NewEma(name string, period int) *EMA {
	return &EMA{
		Name:   name,
		Window: period,
		Type:   "EMA",
	}
}

func (ema *EMA) InitialCalc(candles *data.Candles) {
	var emaPrev float64

	for i := range *candles {
		if i > ema.Window {
			// Prepare a sample-window of data for EMA
			sampleOfCandles := (*candles)[i-ema.Window+1 : i+1]

			// Include only required values
			prices := make([]float64, ema.Window)
			for j, c := range sampleOfCandles {
				prices[j] = c.Close
			}

			// Calculate EMA and record value to indicator's column of a candle
			emaVal := ema.Calc(&prices, emaPrev)
			(*candles)[i].TA[ema.Name] = append((*candles)[i].TA[ema.Name], emaVal)

			emaPrev = emaVal
		}
	}
}

// CalcEMA calculates EMA without instance of EMA struct
func CalcEMA(src *[]float64, prevEma float64) float64 {
	var emaVal float64
	if len(*src) == 0 {
		log.Fatal("RMA: provided data is empty")
	}
	length := float64(len(*src))

	alpha := 2 / (length + 1)

	for _, source := range *src {
		if prevEma == 0 {
			emaVal = SMA(src)
		} else {
			emaVal = alpha*source + (1-alpha)*prevEma
		}
	}

	return emaVal
}

// Calc EMA = alpha * source + (1 - alpha) * prevEma, where alpha = 2 / (length + 1)
func (ema *EMA) Calc(src *[]float64, prevEma float64) float64 {
	return CalcEMA(src, prevEma)
}

func (ema *EMA) GetName() string {
	return ema.Name
}
func (ema *EMA) GetType() string {
	return ema.Type
}
