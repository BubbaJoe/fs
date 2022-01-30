/*
Copyright © 2022 Joe Williams <joewilliamsis@live.com>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteFileCmd represents the deleteFile command
var deleteFileCmd = &cobra.Command{
	Use:   "deleteFile",
	Short: "A brief description of your command",
	Long:  `A longer description that spans multiple`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deleteFile called")
	},
}

func init() {
	rootCmd.AddCommand(deleteFileCmd)

	// Server URL
	deleteFileCmd.Flags().StringP("url", "u", "http://localhost:8080", "url for connecting to the server")

	// Auto Confirm
	deleteFileCmd.Flags().BoolP("auto-confirm", "a", false, "automatically confirm delete prompt")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteFileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteFileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
