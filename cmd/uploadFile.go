/*
Copyright © 2022 Joe Williams <joewilliamsis@live.com>

*/
package cmd

import (
	"fmt"
	"fs-store/client"
	"os"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

// uploadFileCmd represents the uploadFile command
var uploadFileCmd = &cobra.Command{
	Use:   "upload",
	Short: "upload a file from the file system",
	Long:  "upload a file from the file system",
	Run: func(cmd *cobra.Command, paths []string) {
		// Configure the client
		serverUrl := cmd.Flag("url").Value.String()
		client, err := client.NewFSClientConfig(serverUrl, false)
		if err != nil {
			fmt.Println("error: " + err.Error())
			return
		}

		// Check whether to overwrite flag is set/valid
		overwrite, err := cmd.Flags().GetBool("overwrite")
		if err != nil {
			overwriteVal := cmd.Flag("overwrite").Value.String()
			fmt.Printf("invalid value for overwrite flag: '%s'\n", overwriteVal)
			return
		}

		// Find duplicate strings in args array
		var dupes []string
		var seen = make(map[string]struct{})
		for _, path := range paths {
			if _, ok := seen[path]; ok {
				dupes = append(dupes, path)
			}
		}

		// Check for duplicate file names
		if len(dupes) > 0 {
			fmt.Printf("duplicate file paths found: %s\n", strings.Join(dupes, ", "))
			return
		}
		// tmpl := `{{ cyan "Uploading:" }} {{ bar . "[" "-" "->" "." "]"}} {{total . }} {{percent .}}`

		fmt.Println()
		// Upload the files specified in the paths (args)
		for _, path := range paths {
			fmt.Println("Uploading file: '" + path + "' to " + client.Client.BaseURL)
			file, err := os.Open(path)
			if err != nil {
				fmt.Println("error: " + err.Error())
				continue
			}
			defer file.Close()

			// Get the file size
			info, err := file.Stat()
			if err != nil {
				fmt.Println("error: " + err.Error())
				continue
			}
			fileSize := info.Size()

			// bar := progressbar.NewOptions64(fileSize,
			// 	progressbar.OptionEnableColorCodes(true),
			// 	progressbar.OptionShowBytes(true),
			// 	progressbar.OptionThrottle(5*time.Millisecond),
			// 	progressbar.OptionShowCount(),
			// 	progressbar.OptionSetDescription(fmt.Sprintf("[cyan][%d/%d][reset] Uploading file '%s'...\n", i+1, len(paths), path)),
			// 	progressbar.OptionSetTheme(progressbar.Theme{
			// 		Saucer:        "[green]=[reset]",
			// 		SaucerHead:    "[green]>[reset]",
			// 		SaucerPadding: " ",
			// 		BarStart:      "[",
			// 		BarEnd:        "]",
			// 	}))

			// barReader := progressbar.NewReader(file, bar)

			bar := pb.Default.Start64(fileSize)
			// bar.SetTemplate(pb.ProgressBarTemplate{

			// })
			// create proxy reader
			barReader := bar.NewProxyReader(file)
			bar.SetRefreshRate(time.Millisecond * 10)

			err = client.UploadFile(file.Name(), barReader.Reader, overwrite)
			if err != nil {
				fmt.Println("error: " + err.Error())
				return
			}

			bar.Finish()
			fmt.Println("")
		}
	},
}

func init() {
	rootCmd.AddCommand(uploadFileCmd)

	// Auto Confirm
	uploadFileCmd.Flags().BoolP("auto-confirm", "a", false, "automatically confirm upload prompt")

	// Overwrite
	uploadFileCmd.Flags().BoolP("overwrite", "o", false, "overwrite existing file")

	// Server URL
	uploadFileCmd.Flags().StringP("url", "u", "http://localhost:8080", "url for connecting to the server")
}
