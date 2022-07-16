package cmd

import (
	"github.com/janrockdev/crypto-desk/connector"
	"github.com/spf13/cobra"
)

var (
	cmdAlpaca = &cobra.Command{
		Use:   "alpaca",
		Short: "Pricing API - Alpaca",
		RunE:  runAlpaca,
	}
)

func init() {
	rootCmd.AddCommand(cmdAlpaca)
}

func runAlpaca(cmd *cobra.Command, args []string) error {
	return connector.RunAlpaca()
}
