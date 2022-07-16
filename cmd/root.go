package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "crypto-desk",
	Short: "--------------------------------------------------------------------\n" +
		"CRYPTO-DESK PLATFORM\n" +
		"--------------------------------------------------------------------\n" +
		"Description: \n" +
		"Version: 1.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
