package connector

import (
	"context"
	"flag"
	"fmt"
	kdb "github.com/sv/kdbgo"
	"os"
	"sync"
	"time"

	"github.com/janrockdev/crypto-desk/common"

	"github.com/alpacahq/alpaca-trade-api-go/v2/marketdata/stream"
)

var apiKeyPR string
var apiSecretPR string

func init() {
	apiKeyPR = os.Getenv("ALPACAKEY")
	apiSecretPR = os.Getenv("ALPACASEC")
}

func insertPrices(con *kdb.KDBConn, tbl string, tstamp time.Time, symbol string, bidprice float64, askprice float64, midprice float64, exch string) {
	common.Logr.Info(tbl, symbol, tstamp, bidprice, askprice, midprice, exch)
	ts := &kdb.K{kdb.KP, kdb.NONE, []time.Time{tstamp}}
	source := &kdb.K{kdb.KS, kdb.NONE, []string{"alpaca-quotes"}}
	sym := &kdb.K{kdb.KS, kdb.NONE, []string{symbol}}
	bid := &kdb.K{kdb.KF, kdb.NONE, []float64{bidprice}}
	ask := &kdb.K{kdb.KF, kdb.NONE, []float64{askprice}}
	mid := &kdb.K{kdb.KF, kdb.NONE, []float64{midprice}}
	ex := &kdb.K{kdb.KS, kdb.NONE, []string{exch}}
	tab := &kdb.K{kdb.XT, kdb.NONE, kdb.Table{[]string{"ts", "source", "sym", "bid", "mid", "ask", "exchange"}, []*kdb.K{ts, source, sym, bid, mid, ask, ex}}}
	// insert tab sync
	_, err := con.Call("insert", &kdb.K{-kdb.KS, kdb.NONE, tbl}, tab)
	if err != nil {
		fmt.Println("Insert Query failed:", err)
		return
	}
}

func startAlpacaPR() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	con, err := kdb.DialKDB("localhost", 5555, "")
	if err != nil {
		fmt.Println("KDB dial failed:", err)
		return
	}

	logr.Info("Alpaca.markerts pricing consumer is running...")

	c := stream.NewCryptoClient(
		stream.WithLogger(&logger{}),
		stream.WithCredentials(apiKeyPR, apiSecretPR),
		stream.WithCryptoQuotes(func(cq stream.CryptoQuote) {
			fmt.Printf("QUOTE: %+v\n", cq)
			midPrice := (cq.BidPrice + cq.AskPrice) / 2
			insertPrices(con, "quotes", cq.Timestamp, cq.Symbol, cq.BidPrice, cq.AskPrice, midPrice, cq.Exchange)
		}, "BTCUSD"),
	)
	if err := c.Connect(ctx); err != nil {
		panic(err)
	}
	if err := <-c.Terminated(); err != nil {
		panic(err)
	}
}

type logger struct{}

func (l *logger) Infof(format string, v ...interface{}) {
	logr.Println(fmt.Sprintf("INFO "+format, v...))
}

func (l *logger) Warnf(format string, v ...interface{}) {
	logr.Println(fmt.Sprintf("WARN "+format, v...))
}

func (l *logger) Errorf(format string, v ...interface{}) {
	logr.Println(fmt.Sprintf("ERROR "+format, v...))
}

func runAlpacaPR(
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
) error {
	startAlpacaPR()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
	}()

	return nil
}

func RunAlpacaPR() error {
	logr.Infof("Starting Alpaca...")
	flag.Parse()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	if err := runAlpacaPR(cancel, &wg); err != nil {
		return fmt.Errorf("error when starting server: %v", err)
	}
	wg.Wait()

	return nil
}
