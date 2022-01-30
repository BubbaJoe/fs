package cmd

import (
	"errors"
	"fs-store/server"
	"strconv"

	"github.com/spf13/cobra"
)

// startServerCmd represents the startServer command
var startServerCmd = &cobra.Command{
	Use:   "server",
	Short: "starts the FS-Store server",
	RunE: func(cmd *cobra.Command, args []string) error {
		host := cmd.Flag("host").Value.String()
		port := cmd.Flag("port").Value.String()
		dataDir := cmd.Flag("data-dir").Value.String()
		logLevel := cmd.Flag("log-level").Value.String()

		maxFileSizeMBStr := cmd.Flag("max-mb").Value.String()
		maxFileSizeMB, err := strconv.ParseInt(maxFileSizeMBStr, 10, 64)
		if err != nil {
			return errors.New("max-mb must be an integer")
		}
		server, err := server.NewServerConfig(host+":"+port,
			dataDir, 1024*1024*maxFileSizeMB, logLevel)
		if err != nil {
			return err
		}
		return server.StartServer()
	},
}

func init() {
	rootCmd.AddCommand(startServerCmd)

	// Log Level
	startServerCmd.Flags().StringP("log-level", "l", "info", "log level")

	// Host
	startServerCmd.Flags().StringP("host", "H", "127.0.0.1", "host for the server")

	// Port
	startServerCmd.Flags().IntP("port", "p", 8080, "port for the server")

	// Data Directory
	startServerCmd.Flags().StringP("data-dir", "d", "./", "data directory for the server")

	// Max File Size in MB
	startServerCmd.Flags().Int64P("max-mb", "m", 1024, "max file size in MB")

}
