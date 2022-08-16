package connector

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v2/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v2/marketdata"

	"github.com/janrockdev/crypto-desk/indicators"
	"github.com/janrockdev/crypto-desk/indicators/utils"

	"github.com/janrockdev/crypto-desk/common"
	"github.com/shopspring/decimal"

	kdb "github.com/sv/kdbgo"

	lr "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type longShortAlgo struct {
	tradeClient alpaca.Client
	dataClient  marketdata.Client
}

var algo longShortAlgo

func init() {
	apiKey := "PK77HGS61VVXY8XREZ51"
	apiSecret := "4RRfbcX63sJ0cfLQ3y09B6bi6wehr32SxS66qgR4"
	baseURL := "https://paper-api.alpaca.markets"

	algo = longShortAlgo{
		tradeClient: alpaca.NewClient(alpaca.ClientOpts{
			ApiKey:    apiKey,
			ApiSecret: apiSecret,
			BaseURL:   baseURL,
		}),
		dataClient: marketdata.NewClient(marketdata.ClientOpts{
			ApiKey:    apiKey,
			ApiSecret: apiSecret,
		}),
	}
}

type result struct {
	hlc float64
	ema []float64
	rsi []float64
	bb  []float64
}

type tval struct {
	price  float64
	size   int
	status string
	flag   string
}

// type Transactions struct {
// 	Timestamp string
// 	Book      string
// 	Value     string
// }

// // RSI params //TODO: ML adjustment
var rsiup = 68.0
var rsilo = 40.0

// // Take Profit / Stop Loss //TODO: ML adjustment
// var tp = float64(2.8)
// var sl = float64(1.3)

var r []result

// var signal string
var t map[time.Time]tval

var logr = &lr.Logger{
	Out:   os.Stdout,
	Level: lr.InfoLevel,
	Formatter: &prefixed.TextFormatter{
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	},
}

func toStruct(tbl kdb.Table) []utils.OHLCV {

	var data = []utils.OHLCV{}

	nrows := int(tbl.Data[0].Len())
	for i := 0; i < nrows; i++ {
		rec := utils.OHLCV{Time: tbl.Data[0].Index(i).(time.Time), Open: tbl.Data[1].Index(i).(float64), High: tbl.Data[2].Index(i).(float64), Low: tbl.Data[3].Index(i).(float64), Close: tbl.Data[4].Index(i).(float64), Volume: tbl.Data[5].Index(i).(float64)}
		data = append(data, rec)
	}
	return data
}

// func round(num float64) int {
// 	return int(num + math.Copysign(0.5, num))
// }

// func toFixed(num float64, precision int) float64 {
// 	output := math.Pow(10, float64(precision))
// 	return float64(round(num*output)) / output
// }

func (alp longShortAlgo) submitMarketOrder(qty int, symbol string, side string) error {
	account, err := algo.tradeClient.GetAccount()
	if err != nil {
		common.Logr.Errorf("get account: %w", err)
		return err
	}
	if qty > 0 {
		adjSide := alpaca.Side(side)
		decimalQty := decimal.NewFromInt(int64(10))
		_, err := algo.tradeClient.PlaceOrder(alpaca.PlaceOrderRequest{
			AccountID:   account.ID,
			AssetKey:    &symbol,
			Qty:         &decimalQty,
			Side:        adjSide,
			Type:        "market",
			TimeInForce: "day",
		})
		if err == nil {
			common.Logr.Infof("Market order of | %d %s %s | completed\n", qty, symbol, side)
		} else {
			common.Logr.Errorf("Order of | %d %s %s | did not go through: %s.", qty, symbol, side, err)
		}
		return err
	}
	common.Logr.Infof("Quantity is <= 0, order of | %d %s %s | not sent.", qty, symbol, side)
	return nil
}

func main() {
	t = make(map[time.Time]tval)

	//key := os.Getenv("OANDA_API_KEY")
	//accountID := os.Getenv("OANDA_ACCOUNT_ID")
	//oanda := goanda.NewConnection(accountID, key, false)
	con, err := kdb.DialKDB("localhost", 5000, "")

	ticker := time.NewTicker(5 * time.Second)
	done := make(chan bool)
	var res []*kdb.K

	go func() {

		for {
			select {
			case <-done:
				return
			case _ = <-ticker.C:

				_, err = con.Call("delete strategy from `.")
				if err != nil {
					fmt.Println("Insert Query failed:", err)
					return
				}

				ktbl, err := con.Call("select [-36000] ts, bid, ask, mid, mid, mid from quotes")
				if err != nil {
					fmt.Println("Query failed:", err)
					return
				}

				series := toStruct(ktbl.Data.(kdb.Table))

				indicRSI := indicators.NewRSI(1000, "close")
				var tempResRSI []float64

				indicBB := indicators.NewBB(6000, "close")
				var tempResBB []float64
				var flag string

				for _, v := range series {
					tempResRSI = indicRSI.Calculate(v)
					tempResBB = indicBB.Calculate(v)

					r = append(r, result{
						hlc: v.HLC3(),
						rsi: tempResRSI,
						bb:  tempResBB})
					colTS := &kdb.K{kdb.KP, kdb.NONE, []time.Time{v.Time.Local()}}
					colSYM := &kdb.K{kdb.KS, kdb.NONE, []string{"SOLUSD"}}
					colHLC := &kdb.K{kdb.KF, kdb.NONE, []float64{v.HLC3()}}
					colRSI := &kdb.K{kdb.KF, kdb.NONE, []float64{50.0}}
					if !math.IsNaN(tempResRSI[0]) {
						colRSI = &kdb.K{kdb.KF, kdb.NONE, []float64{tempResRSI[0]}}
					}
					// colBBb := &kdb.K{kdb.KF, kdb.NONE, []float64{v.HLC3()}}
					// colBBu := &kdb.K{kdb.KF, kdb.NONE, []float64{v.HLC3()}}
					// colBBl := &kdb.K{kdb.KF, kdb.NONE, []float64{v.HLC3()}}
					//colSize := &kdb.K{kdb.KF, kdb.NONE, []float64{1.0}}
					colBuy := &kdb.K{kdb.KF, kdb.NONE, []float64{math.NaN()}}
					colSell := &kdb.K{kdb.KF, kdb.NONE, []float64{math.NaN()}}
					flag = ""
					if tempResBB[0] > 0 {
						// colBBb = &kdb.K{kdb.KF, kdb.NONE, []float64{tempResBB[0]}}
						// colBBu = &kdb.K{kdb.KF, kdb.NONE, []float64{tempResBB[1]}}
						// colBBl = &kdb.K{kdb.KF, kdb.NONE, []float64{tempResBB[2]}}

						if v.GetByType("close") < tempResBB[2] {
							if rsilo > tempResRSI[0] {
								//logr.Infof("%v > %v", rsilo, tempResRSI[0])
								colBuy = &kdb.K{kdb.KF, kdb.NONE, []float64{v.HLC3()}}
								flag = "buy"
								//toTrade(*con, "buy", v.Time, v.HLC3())
								//logr.Info("buy", v.Time)
								//algo.submitMarketOrder(10, "SOLUSD", "buy")
							}
						}
						if v.GetByType("close") > tempResBB[1] {
							if rsiup < tempResRSI[0] {
								//logr.Infof("%v < %v", rsiup, tempResRSI[0])
								colSell = &kdb.K{kdb.KF, kdb.NONE, []float64{v.HLC3()}}
								flag = "sell"
								//logr.Info("sell", v.Time, v.HLC3())
								//toTrade(*con, "sell", v.Time, v.HLC3())
							}
						}
					}
					//res = []*kdb.K{colTS, colSYM, colHLC, colRSI, colBBb, colBBu, colBBl, colBuy, colSell}
					res = []*kdb.K{colTS, colSYM, colHLC, colRSI, colBuy, colSell}
					// tab := &kdb.K{kdb.XT, kdb.NONE, kdb.Table{[]string{"ts", "sym", "hlc", "ema", "rsi", "bbb", "bbu", "bbl", "buy", "sell"},
					// 	[]*kdb.K{colTS, colSYM, colHLC, colEMA, colRSI, colBBb, colBBu, colBBl, colBuy, colSell}}}
					// // insert tab sync
					// _, err = con.Call("insert", &kdb.K{-kdb.KS, kdb.NONE, "strategy"}, tab)
					// if err != nil {
					// 	fmt.Println("Insert Query failed:", err)
					// 	return
					// }
				}
				// reset(GC)
				r = []result{}
				if flag != "" {
					logr.Infof("\n%v (detail:%v)", flag, res)
					if flag == "buy" {
						algo.submitMarketOrder(10, "SOLUSD", "buy")
					}
					if flag == "sell" {
						//logr.Info("Sell")
						algo.submitMarketOrder(10, "SOLUSD", "sell")
					}
				}
				//} else {
				//	//fmt.Print(".")
				//	positions := positions.PositionsAlpaca()
				//	for v, position := range positions {
				//		fmt.Println(v, position.Symbol, position.Qty, position.Side, position.CurrentPrice, position.EntryPrice, position.UnrealizedPL)
				//	}
				//}
			}
		}
	}()
	for {
		time.Sleep(23 * time.Hour)
		err := con.Close()
		if err != nil {
			return
		}
	}
}
