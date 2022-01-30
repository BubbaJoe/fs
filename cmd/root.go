package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fs-store",
	Short: "A fast file store for storing files remotely",
	SilenceUsage: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func setupCommonClientFlags(cmd *cobra.Command) {
	// Server URL
	cmd.Flags().StringP("url", "u", "http://localhost:8080",
		"url for connecting to the server")

	// Verbose
	cmd.Flags().BoolP("verbose", "v", false, "verbose output")
}
