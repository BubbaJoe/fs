/*
Copyright © 2022 Joe Williams <joewilliamsis@live.com>

*/
package cmd

import (
	"fmt"
	"fs-store/client"
	"strings"

	"github.com/spf13/cobra"
)

// listFilesCmd represents the listFile command
var listFilesCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long:  `A lon`,
	Run: func(cmd *cobra.Command, args []string) {
		serverUrl := cmd.Flag("url").Value.String()
		client, err := client.NewFSClientConfig(serverUrl, false)
		if err != nil {
			fmt.Println(err)
		}

		files, err := client.ListFiles()
		if err != nil {
			fmt.Println(err)
		}

		// Print the files
		fmt.Print("Listing Files: ")
		if len(files) != 0 {
			// write file names
			var fileNames []string
			for _, file := range files {
				fileNames = append(fileNames, file.FileName)
			}
			fmt.Println(strings.Join(fileNames, ", "))

		} else {
			fmt.Println("No Files Found")
		}
	},
}

func init() {
	rootCmd.AddCommand(listFilesCmd)

	// Server URL
	listFilesCmd.Flags().StringP("url", "u", "http://localhost:8080", "url for connecting to the server")

	// Verbose
	listFilesCmd.Flags().BoolP("verbose", "v", false, "verbose output")
}
