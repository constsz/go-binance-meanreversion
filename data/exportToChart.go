package data

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
)

type Table struct {
	Headers Headers
	Rows    [][]string
}
type Headers []Header

type Header struct {
	Name string
	Type string
}

// ExportToChart runs several exporting visualization functions:
// - export candles and indicators
// - export backtesting results as a table
// Currenty EVERYTHING works only for 1 strategy, 1 symbol, 1 tf. This will be improved in future.
//
//	indicators: [ [Name, Type], [Name, Type] ]
func (candles *Candles) ExportToChart(candleLimit int, indicators [][]string) {
	// Table
	var table Table

	// Set maxNumberOfCandlesAllowed to keep frontend from freezing
	maxNumberOfCandlesAllowed := 1000
	if candleLimit == 0 {
		if len(*candles) > maxNumberOfCandlesAllowed {
			candleLimit = maxNumberOfCandlesAllowed
		}
	}

	// Headers
	//"Timestamp", "Open", "High", "Low", "Close"
	table.Headers = Headers{
		Header{"Timestamp", "ohlcv"},
		Header{"Open", "ohlcv"},
		Header{"High", "ohlcv"},
		Header{"Low", "ohlcv"},
		Header{"Close", "ohlcv"},
	}

	// Indicators List
	// First we get "indicatorsList" separately,
	// and then append it to the main table.Headers
	var indicatorsHeaders []Header

	for key := range (*candles)[len(*candles)-1].TA {
		if indicators != nil {
			for _, v := range indicators {
				if key == v[0] {
					// Depending on type of indicator, it receives one or more columns
					switch v[1] {
					// BetterBands has 7 parameters
					case "BetterBands":
						indicatorValues := [3]string{"LowBand", "HighBand", "Regime"}
						for _, iv := range indicatorValues {
							indicatorName := v[0]
							indicatorType := v[1] + "_" + iv

							indicatorHeader := Header{
								indicatorName, indicatorType,
							}
							indicatorsHeaders = append(indicatorsHeaders, indicatorHeader)
						}
					case "TrendMove":
						indicatorValues := [2]string{"TrendLine", "MoveLine"}
						for _, iv := range indicatorValues {
							indicatorName := v[0]
							indicatorType := v[1] + "_" + iv

							indicatorHeader := Header{
								indicatorName, indicatorType,
							}
							indicatorsHeaders = append(indicatorsHeaders, indicatorHeader)
						}
					// Default for Indicators with 1 parameter
					default:
						indicatorName := v[0]
						indicatorType := v[1]

						indicatorHeader := Header{
							indicatorName, indicatorType,
						}
						indicatorsHeaders = append(indicatorsHeaders, indicatorHeader)
					}
				}
			}
		}
	}

	table.Headers = append(table.Headers, indicatorsHeaders...)

	var candlesSlice Candles

	if candleLimit == 0 {
		// If no candleLimit - using all candles
		if ((*candles)[len(*candles)-1].Counter) != (*candles)[len(*candles)-2].Counter {
			candlesSlice = (*candles)[:len(*candles)-1]
		} else {
			candlesSlice = *candles
		}

		// If candleLimit is present - cut only last n-number of candles
	} else {
		candlesSlice = (*candles)[len(*candles)-candleLimit:]
	}

	// ---
	// Rows and Columns
	for _, c := range candlesSlice {
		// basic Row data
		row := []string{
			strconv.FormatInt(c.Time, 10),
			fmt.Sprintf("%f", c.Open),
			fmt.Sprintf("%f", c.High),
			fmt.Sprintf("%f", c.Low),
			fmt.Sprintf("%f", c.Close),
		}

		// add indicators to Row
		for _, i := range indicatorsHeaders {
			if len(c.TA[i.Name]) > 0 {
				switch i.Type {
				// BetterBands:
				// [1: LowBand, 2: HighBand, 3: Regime (long/short)]
				case "BetterBands_LowBand":
					row = append(row, fmt.Sprintf("%f", c.TA[i.Name][1]))
				case "BetterBands_HighBand":
					row = append(row, fmt.Sprintf("%f", c.TA[i.Name][2]))
				case "BetterBands_Regime":
					row = append(row, fmt.Sprintf("%f", c.TA[i.Name][3]))
				case "TrendMove_TrendLine":
					row = append(row, fmt.Sprintf("%f", c.TA[i.Name][0]))
				case "TrendMove_MoveLine":
					row = append(row, fmt.Sprintf("%f", c.TA[i.Name][1]))
				default:
					row = append(row, fmt.Sprintf("%f", c.TA[i.Name][0]))
				}

			}
		}

		// append Row to the Table
		table.Rows = append(table.Rows, row)
	}

	// File Names
	fileName_Candles := "candles"
	fileName_Headers := "headers"

	// WRITE Csv to Disk : Candles
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(path.Join(workingDir, fmt.Sprintf("chart/html/data/%v.csv", fileName_Candles)))
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("exportToChart: Error during creating FILE: chart/html/data/%v.csv", fileName_Candles)
		}
	}()
	if err != nil {
		log.Fatal(err)
	}

	w := csv.NewWriter(f)
	err = w.WriteAll(table.Rows) // calls Flush internally
	if err != nil {
		log.Fatal(err)
	}

	// WRITE Json to Disk : Headers
	file, err := json.MarshalIndent(table.Headers, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(path.Join(workingDir, fmt.Sprintf("chart/html/data/%v.json", fileName_Headers)), file, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nExported to chart.\n")
}
