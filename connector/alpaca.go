package connector

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
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
	baseURL = os.Getenv("ALPACAAPI")
	apiKey = os.Getenv("ALPACAKEY")
	apiSecret = os.Getenv("ALPACASEC")
}

func percentageChange(old, new float64) (delta float64) {
	return ((new - old) / old) * 100
}

func makeSound(size float64, side string) {
	i := math.Round(size + 1)
	level := fmt.Sprintf("set Volume %v", i)
	//cmd := exec.Command("say", "-v", "Susan", "buy", s)
	if side == "B" {
		cmd := exec.Command("osascript", "-e", level)
		if err := cmd.Run(); err != nil {
			common.Logr.Fatal(err)
		}
		cmd = exec.Command("afplay", "/System/Library/Sounds/Funk.aiff")
		if err := cmd.Run(); err != nil {
			common.Logr.Fatal(err)
		}
	} else {
		cmd := exec.Command("osascript", "-e", level)
		if err := cmd.Run(); err != nil {
			common.Logr.Fatal(err)
		}
		cmd = exec.Command("afplay", "/System/Library/Sounds/Ping.aiff")
		if err := cmd.Run(); err != nil {
			common.Logr.Fatal(err)
		}
	}
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
				if ct.Size >= 0.1 {
					go makeSound(ct.Size, "B")
				}
			} else {
				res = result{ct.Timestamp.String(), ct.Price, mark, res.suma - ct.Size}
				if ct.Size >= 0.1 {
					go makeSound(ct.Size, "S")
				}
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
