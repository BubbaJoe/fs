/*
Copyright © 2022 Joe Williams <joewilliamsis@live.com>

*/
package cmd

import (
	"fmt"
	"fs-store/server"

	"github.com/spf13/cobra"
)

// startServerCmd represents the startServer command
var startServerCmd = &cobra.Command{
	Use:   "start-server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Starting File System Storage (fs-store) Server")
		dataDir := cmd.Flag("data-dir").Value.String()
		host := cmd.Flag("host").Value.String()
		port := cmd.Flag("port").Value.String()

		fmt.Println("data directory: ", dataDir)
		fmt.Println("host: ", host)
		fmt.Println("port: ", port)
		server, err := server.NewServerConfig(host+":"+port,
			dataDir, 1<<32-1, 128, true)
		if err != nil {
			return err
		}
		return server.StartServer()
	},
}

func init() {
	rootCmd.AddCommand(startServerCmd)

	// Host
	startServerCmd.Flags().StringP("host", "H", "127.0.0.1", "host for the server")

	// Port
	startServerCmd.Flags().IntP("port", "p", 8080, "port for the server")

	// Data Directory
	startServerCmd.Flags().StringP("data-dir", "d", "./", "data directory for the server")
}
