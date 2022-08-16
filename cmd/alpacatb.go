package cmd

import (
	"github.com/janrockdev/crypto-desk/connector"
	"github.com/spf13/cobra"
)

var (
	cmdAlpacaTB = &cobra.Command{
		Use:   "alpacatb",
		Short: "Alpaca.Markets Exchange TradeBook",
		RunE:  runAlpacaTB,
	}
)

func init() {
	rootCmd.AddCommand(cmdAlpacaTB)
}

func runAlpacaTB(cmd *cobra.Command, args []string) error {
	return connector.RunAlpacaTB()
}
