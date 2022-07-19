package indicators

import (
	"github.com/janrockdev/crypto-desk/indicators/utils"
)

//Indicator : Interface for all indicators
type Indicator interface {
	Calculate(newData utils.OHLCV) []float64
}
