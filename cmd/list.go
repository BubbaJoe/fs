package cmd

import (
	"errors"
	"fmt"
	"fs-store/client"
	"strings"

	"github.com/spf13/cobra"
)

// listFilesCmd represents the listFile command
var listFilesCmd = &cobra.Command{
	Use:   "list",
	Short: "list files on the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverUrl := cmd.Flag("url").Value.String()
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		client, err := client.NewFSClientConfig(serverUrl, verbose)
		if err != nil {
			return err
		}

		files, err := client.ListFiles()
		if err != nil {
			return err
		} else if files == nil {
			return errors.New("could fetch files from server")
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
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listFilesCmd)
	setupCommonClientFlags(listFilesCmd)
}
