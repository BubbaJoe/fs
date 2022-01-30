package cmd

import (
	"errors"
	"fmt"
	"fs-store/client"

	"github.com/spf13/cobra"
)

// deleteFileCmd represents the deleteFile command
var deleteFileCmd = &cobra.Command{
	Use:   "delete [file] [?file2] ...",
	Short: "delete a file from the server",
	Args: func(cmd *cobra.Command, paths []string) error {
		// Find duplicate strings in paths
		var seen = make(map[string]struct{})
		for _, path := range paths {
			if _, ok := seen[path]; ok {
				return errors.New("duplicate file path: " + path)
			} else {
				seen[path] = struct{}{}
			}
		}

		return cobra.MinimumNArgs(1)(cmd, paths)
	},
	RunE: func(cmd *cobra.Command, paths []string) error {
		// Configure the client
		serverUrl := cmd.Flag("url").Value.String()
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		client, err := client.NewFSClientConfig(serverUrl, verbose)
		if err != nil {
			return err
		}

		// Upload the files specified in the paths (args)
		for _, path := range paths {
			fmt.Println("Deleting file: '" + path + "' from " + client.Client.BaseURL)

			if err := client.DeleteFile(path); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteFileCmd)
	setupCommonClientFlags(deleteFileCmd)
}
