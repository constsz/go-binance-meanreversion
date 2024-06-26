package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type SymbolsSettings []SymbolSettings

type SymbolSettings struct {
	Symbol            string
	Quantity          float64
	TickSize          string
	QuantityPrecision int
}

type SymbolsSettingsJson []SymbolSettingsJson

type SymbolSettingsJson struct {
	Symbol string  `json:"Symbol"`
	QtyUSD float64 `json:"QtyUSD"`
}

func (sss *SymbolsSettings) SymbolTickSize(symbol string) string {
	var symbolTickSizeString string
	for _, settings := range *sss {
		if settings.Symbol == symbol {
			symbolTickSizeString = settings.TickSize
		}
	}
	return symbolTickSizeString
}

func readSymbolSettingsJson() *SymbolsSettingsJson {
	filePath := "state/symbolsSettings.json"
	_, err := os.Stat(filePath)
	if err != nil {
		log.Fatal("readSymbolSettingsJson: Can't find JSON file on that path.\n", err)
	} else {
		//	Read file and return it's contents
		f, err := os.Open(filePath)
		if err != nil {
			log.Fatal("readSymbolSettingsJson: error reading Settings.json file!")
		}
		defer f.Close()

		// Unmarshall json to struct
		bytes, _ := io.ReadAll(f)

		var ssj SymbolsSettingsJson

		err = json.Unmarshal(bytes, &ssj)

		if err != nil {
			log.Fatal("readSymbolSettingsJson: error during Unmarshalling the json contents.")
		}

		return &ssj
	}

	return nil
}

// GetSymbolSettings gets only settings for currently used symbols from main list.
func GetSymbolSettings(client *futures.Client, symbolsList []string) SymbolsSettings {
	// client.ExchangeInfo ... чтобы получить все Precisions.
	res, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Initialize main variable
	var symbolSettings SymbolsSettings

	// Read SymbolJson settings
	ssJson := readSymbolSettingsJson()

	var defaultSettings SymbolSettings

	for _, ssj := range *ssJson {
		if ssj.Symbol == "DEFAULT" {
			defaultSettings.Symbol = ssj.Symbol
			defaultSettings.Quantity = ssj.QtyUSD
		}
	}

	for _, symbol := range symbolsList {
		var symbolFound bool
		// For each symbol from my list create the SymbolSettings
		var ss SymbolSettings

		ss.Symbol = symbol

		// Temporary write QtyUSD to Quantity
		for _, ssj := range *ssJson {
			if SymbolUSDT(ssj.Symbol) == symbol {
				ss.Quantity = ssj.QtyUSD
				symbolFound = true
			}
		}

		if !symbolFound {
			ss.Quantity = defaultSettings.Quantity
			fmt.Printf("SymbolSettingsJSON has NO settings for %s! Using DEFAULT.\n", symbol)
		}

		// Fill the Symbol name and the TickSize
		for _, s := range res.Symbols {
			if s.Symbol == symbol {
				ss.TickSize = s.PriceFilter().TickSize
				ss.QuantityPrecision = s.QuantityPrecision
			}
		}

		// Convert Quantity to Symbol (base asset price, instead of USD)
		// Request latest symbol price
		symbolPrices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
		if err != nil {
			log.Fatal("GetSymbolSettings: Error during calling Binance PriceService!", err)
		}
		symbolPrice, _ := strconv.ParseFloat(symbolPrices[0].Price, 64)

		// Convert Quantity from USD to Symbol quantity
		ss.Quantity = ConvertQuantityFromUSDT(ss.Quantity, symbolPrice, ss.QuantityPrecision)

		symbolSettings = append(symbolSettings, ss)

	}
	// Request tickers from Binance
	// For each symbol: convert QtyUSD to base asset Qty and rewrite Quantity field in SymbolSettings struct

	return symbolSettings
}

// Settings represent global system-wide settings.
// Json parsed right here.
type Settings struct {
	Symbols                []string `json:"Symbols"`
	CandleTickPeriod       int      `json:"CandleTickPeriod"`
	OrderTimeSpacing       int      `json:"OrderTimeSpacing"`
	MaxOrderCountPerSymbol int      `json:"MaxOrderCountPerSymbol"`
}

// Initialization functions

func InitSettings() *Settings {
	// Check if file exists
	if _, err := os.Stat("Settings.json"); err != nil {
		// if file not exist: raise a serious error!
		panic("Settings.json file was not found! Can't continue without it!")
	} else {
		//	Read file and return it's contents
		f, err := os.Open("Settings.json")
		if err != nil {
			log.Fatal("storage: error reading Settings.json file!")
		}
		defer f.Close()

		// Unmarshall json to struct
		bytes, _ := io.ReadAll(f)

		var settings Settings

		err = json.Unmarshal(bytes, &settings)

		if err != nil {
			log.Fatal("Settings: error during Unmarshalling the json contents.")
		}

		for i := range settings.Symbols {
			settings.Symbols[i] = settings.Symbols[i]
		}

		return &settings
	}
}

func SymbolUSDT(symbol string) string {
	return symbol + "USDT"
}

func SymbolUSDTList(symbols []string) []string {
	symbolsUSDT := make([]string, len(symbols))

	for i, s := range symbols {
		symbolsUSDT[i] = SymbolUSDT(s)
	}

	return symbolsUSDT

}

func SymbolNoUSDT(symbolUSDT string) string {
	return strings.ReplaceAll(symbolUSDT, "USDT", "")
}

func ConvertQuantityFromUSDT(qtyUSD float64, symbolPrice float64, qtyPrecision int) float64 {
	qtySymbol := RoundToPrecision(qtyUSD/symbolPrice, qtyPrecision)

	return qtySymbol
}

func RoundToPrecision(x float64, precision int) float64 {
	mult := 1.0

	if precision != 0 {
		for i := 0; i < precision; i++ {
			mult *= 10
		}
	}

	x = math.Round(x*mult) / mult

	return x
}

func RoundToPrecisionByString(x float64, precision int) string {
	switch precision {
	case 0:
		return fmt.Sprintf("%.0f", x)
	case 1:
		return fmt.Sprintf("%.1f", x)
	case 2:
		return fmt.Sprintf("%.2f", x)
	case 3:
		return fmt.Sprintf("%.3f", x)
	case 4:
		return fmt.Sprintf("%.4f", x)
	case 5:
		return fmt.Sprintf("%.5f", x)
	case 6:
		return fmt.Sprintf("%.6f", x)
	case 7:
		return fmt.Sprintf("%.7f", x)
	case 8:
		return fmt.Sprintf("%.8f", x)
	case 9:
		return fmt.Sprintf("%.9f", x)
	case 10:
		return fmt.Sprintf("%.10f", x)
	default:
		return fmt.Sprintf("%.4f", x)
	}
}
