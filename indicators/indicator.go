package indicators

import (
	"ligs-project-go/indicators/utils"
)

//Indicator : Interface for all indicators
type Indicator interface {
	Calculate(newData utils.OHLCV) []float64
}
