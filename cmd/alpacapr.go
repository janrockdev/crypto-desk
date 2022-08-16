package cmd

import (
	"github.com/janrockdev/crypto-desk/connector"
	"github.com/spf13/cobra"
)

var (
	cmdAlpacaPR = &cobra.Command{
		Use:   "alpacapr",
		Short: "Alpaca.Markets Exchange Pricing",
		RunE:  runAlpacaPR,
	}
)

func init() {
	rootCmd.AddCommand(cmdAlpacaPR)
}

func runAlpacaPR(cmd *cobra.Command, args []string) error {
	return connector.RunAlpacaPR()
}
