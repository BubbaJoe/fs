package cmd

import (
	"errors"
	"fmt"
	"fs-store/client"
	"os"

	"github.com/spf13/cobra"
)

// uploadFileCmd represents the uploadFile command
var uploadFileCmd = &cobra.Command{
	Use:   "upload [file] [?file2] ...",
	Short: "upload a file from the file system to the server",
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

		// make sure all files exist
		for _, path := range paths {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return errors.New("file does not exist: " + path)
			} else if os.IsPermission(err) {
				return errors.New("permission denied: " + path)
			} else if err != nil {
				return err
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

		// Check whether to overwrite flag is set/valid
		overwrite, err := cmd.Flags().GetBool("overwrite")
		if err != nil {
			overwriteVal := cmd.Flag("overwrite").Value.String()
			fmt.Printf("invalid value for overwrite flag: '%s'\n", overwriteVal)
			return err
		}

		fmt.Println()
		// Upload the files specified in the paths (args)
		for _, path := range paths {
			fmt.Println("Uploading file: '" + path + "' to " + client.Client.BaseURL)
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if err := client.UploadFile(file.Name(), file, overwrite); err != nil {
				return err
			}

			fmt.Println()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uploadFileCmd)
	setupCommonClientFlags(uploadFileCmd)

	// Overwrite
	uploadFileCmd.Flags().BoolP("overwrite", "o", false, "overwrite existing file")
}
