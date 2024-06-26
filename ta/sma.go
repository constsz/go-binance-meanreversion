package ta

import (
	"directionalMaker/data"
	"log"
)

type SmaType struct {
	Name   string
	Window int
	Type   string
}

func NewSMA(name string, period int) *SmaType {
	return &SmaType{
		Name: name, Window: period, Type: "SMA",
	}
}

func SMA(prices *[]float64) float64 {
	if len(*prices) == 0 {
		log.Fatal("SMA: provided data is empty")
	}

	var sum float64
	for _, p := range *prices {
		sum += p
	}
	return sum / float64(len(*prices))
}

func (sma *SmaType) InitialCalc(candles *data.Candles) {
	for i := range *candles {
		if i > sma.Window {
			// Prepare a sample-window of data for SMA
			sampleOfCandles := (*candles)[i-sma.Window+1 : i+1]

			// Include only required values
			prices := make([]float64, sma.Window)
			for j, tick := range sampleOfCandles {
				prices[j] = tick.Close
			}

			// Calculate SMA and record value to indicator's column of a candle
			sma_val := SMA(&prices)
			(*candles)[i].TA[sma.Name] = append((*candles)[i].TA[sma.Name], sma_val)

		}
	}
}

func (sma *SmaType) GetName() string {
	return sma.Name
}
func (sma *SmaType) GetType() string {
	return sma.Type
}
