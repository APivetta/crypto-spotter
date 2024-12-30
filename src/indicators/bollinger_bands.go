package indicators

import (
	"github.com/cinar/indicator"
)

func GetBollingerBands(closePrices []float64) (upperBand, middleBand, lowerBand []float64) {
	// Bollinger Bands with a 20-period SMA and multiplier of 2
	upperBand, middleBand, lowerBand = indicator.BollingerBands(closePrices)

	return
}

// func BollingerBandsAnalysis(prices, upperBand, middleBand, lowerBand, rsi []float64, volume []float64) (*Action, error) {
// 	// Ensure all slices are the same length
// 	if len(prices) == 0 || len(prices) != len(upperBand) || len(prices) != len(lowerBand) || len(prices) != len(middleBand) {
// 		return nil, fmt.Errorf("Error: Data arrays are not of the same length.")
// 	}

// 	bandWidth, _ := indicator.BollingerBandWidth(middleBand, upperBand, lowerBand)
// }
