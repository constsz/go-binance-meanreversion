package data

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func LoadCSVCandles(filePath string) Candles {
	// open file
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Read CSV contents
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	candles := *parseCsvCandles(data)

	return candles

}

func parseCsvCandles(data [][]string) *Candles {
	candles := make(Candles, len(data)-1)

	// Loop over all rows
	for i, row := range data {
		// Process each column in a row
		// omit header row
		if i > 0 {
			var c Candle

			c.TA = make(map[string][]float64, 2)

			for j, field := range row {
				switch j {
				case 0: // Time
					n, err := strconv.ParseInt(field, 10, 64)
					if err != nil {
						log.Fatal("Error parsing Time", err)
					} else {
						c.Time = n
					}
				case 1: // Open
					n, err := strconv.ParseFloat(field, 64)
					if err != nil {
						log.Fatal(err)
					} else {
						c.Open = n
					}
				case 2: // High
					n, err := strconv.ParseFloat(field, 64)
					if err != nil {
						log.Fatal(err)
					} else {
						c.High = n
					}
				case 3: // Low
					n, err := strconv.ParseFloat(field, 64)
					if err != nil {
						log.Fatal(err)
					} else {
						c.Low = n
					}
				case 4: // Close
					n, err := strconv.ParseFloat(field, 64)
					if err != nil {
						log.Fatal(err)
					} else {
						c.Close = n
					}
				case 5: // Quantity
					n, err := strconv.ParseFloat(field, 64)
					if err != nil {
						log.Fatal(err)
					} else {
						c.Quantity = n
					}
				case 6: // VolumeDelta
					n, err := strconv.ParseFloat(field, 64)
					if err != nil {
						log.Fatal(err)
					} else {
						c.VolumeDelta = n
					}
				case 7: // IsClosed
					n, err := strconv.ParseBool(field)
					if err != nil {
						log.Fatal(err)
					} else {
						c.IsClosed = n
					}
				case 8: // Counter
					n, err := strconv.Atoi(field)
					if err != nil {
						log.Fatal("Error parsing Time", err)
					} else {
						c.Counter = n
					}
				}
			}
			// Set the candle Id
			c.Id = i
			// Add candle to all candles
			candles[i-1] = c
		}
	}

	return &candles

}

func ParseDateFromCsvFilename(symbol, fileName string) string {
	replacementString := SymbolUSDT(symbol) + "-candles-"
	dateString := strings.ReplaceAll(fileName, replacementString, "")
	dateString = strings.ReplaceAll(dateString, ".csv", "")

	return dateString
}

func GetListOfSymbols(pathOfSymbolFolders string) []string {
	// list of all folders in a directory
	dirs, err := os.ReadDir(pathOfSymbolFolders)
	if err != nil {
		log.Fatal(err)
	}

	var symbolDirectories []string

	for _, dir := range dirs {
		if dir.IsDir() {
			symbolDirectories = append(symbolDirectories, strings.ReplaceAll(dir.Name(), "USDT", ""))
		}
	}
	fmt.Println(symbolDirectories)
	return symbolDirectories
}
