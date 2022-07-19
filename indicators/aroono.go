package indicators

import (
	"math"

	"github.com/janrockdev/crypto-desk/indicators/utils"
)

//Aroon Oscillator
//Source : https://www.investopedia.com/terms/a/aroonoscillator.asp
//The Aroon Oscillator is a trend-following indicator that uses aspects of the Aroon Indicator (Aroon Up and Aroon Down)
//to gauge the strength of a current trend and the likelihood that it will continue. Readings above zero indicate that
//an uptrend is present, while readings below zero indicate that a downtrend is present. Traders watch for zero line
//crossovers to signal potential trend changes. They also watch for big moves, above 50 or below -50 to signal strong
//price moves.
type aroono struct {
	period    uint
	prev      float64
	priceType string
	buf       *utils.OHLCVBuffer
}

//NewAROONO 	: To return aroono struct instance
//Params
//period		: calculation period
//priceType 	: To use it as base price for calculations from OHLCV Buffer (open, high, low, close)
func NewAROONO(period uint, priceType string) *aroono {
	return &aroono{
		period:    period,
		prev:      math.NaN(),
		priceType: priceType,
		buf:       utils.NewOHLCVBuffer(period + 1),
	}
}

//Calculate : method to Calculate aroono and return results as float array (1 value in array)
//Return	: AroonHigh-AroonLow (1 value in array)
//Params
//newData	: OHLCV instance to use its values for price calculation
func (ins *aroono) Calculate(newData utils.OHLCV) []float64 {
	newPrice := newData.GetByType(ins.priceType)

	if math.IsNaN(newPrice) {
		return []float64{ins.prev}
	}

	ins.buf.Add(newData)

	if ins.buf.Pushes < ins.buf.Size {
		return []float64{math.NaN()}
	}

	aroonHigh := ((float64(ins.period) - (float64(ins.period) - ins.buf.MaxIndex().High)) / float64(ins.period)) * 100.0
	aroonLow := ((float64(ins.period) - (float64(ins.period) - ins.buf.MinIndex().Low)) / float64(ins.period)) * 100.0

	ins.prev = aroonHigh - aroonLow

	return []float64{ins.prev}
}
