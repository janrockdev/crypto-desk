package connector

import (
	"context"
	"flag"
	"fmt"
	"sync"

	"github.com/alpacahq/alpaca-trade-api-go/v2/marketdata/stream"
	"github.com/janrockdev/crypto-desk/common"
)

var apiKey string
var apiSecret string
var baseURL string
var doOnce sync.Once

type result struct {
	timestamp string
	price     float64
	mark      string
	suma      float64
}

func init() {
	apiKey = "PKJH1OV6YQVY7DIHWKAM"
	apiSecret = "T9NgVI4I0KVN5ByxSkicFi3FZ8aKHr7LLFt3hpMy"
	baseURL = "https://paper-api.alpaca.markets"
}

func percentageChange(old, new float64) (delta float64) {
	diff := float64(new - old)
	delta = (diff / float64(old)) * 100
	return
}

func startAlpaca() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var res result
	var prev result
	var mark string
	var start float64

	common.Logr.Info("Alpaca.markerts pricing consumer is running...")

	c := stream.NewCryptoClient(
		stream.WithCredentials(apiKey, apiSecret),
		stream.WithCryptoTrades(func(ct stream.CryptoTrade) {
			doOnce.Do(func() {
				start = ct.Price
			})
			if prev.price > ct.Price {
				mark = fmt.Sprintf("\033[1;31m%s\033[0m", "▼")
			} else {
				mark = fmt.Sprintf("\033[1;32m%s\033[0m", "▲")
			}
			if ct.TakerSide == "B" {
				res = result{ct.Timestamp.String(), ct.Price, mark, res.suma + ct.Size}
			} else {
				res = result{ct.Timestamp.String(), ct.Price, mark, res.suma - ct.Size}
			}
			price := fmt.Sprintf("%.2f", res.price)
			suma := fmt.Sprintf("%.8f", res.suma)
			prc := fmt.Sprintf("%0.2f", percentageChange(start, ct.Price))
			fmt.Printf("\033[2K\r%v: %v %v (%v %v%%) %v", res.timestamp, price, res.mark, start, prc, suma)
			//database.RecInsertOne("quotes", bson.D{{"timestamp", res.timestamp}, {"price", price}})
			prev = res
		}, "BTCUSD"),
	)
	if err := c.Connect(ctx); err != nil {
		panic(err)
	}
	if err := <-c.Terminated(); err != nil {
		panic(err)
	}
}

func runAlpaca(
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
) error {
	startAlpaca()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
	}()

	return nil
}

func RunAlpaca() error {
	common.Logr.Infof("Starting Alpaca...")
	flag.Parse()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	if err := runAlpaca(cancel, &wg); err != nil {
		return fmt.Errorf("Error when starting server: %v", err)
	}
	wg.Wait()

	return nil
}
