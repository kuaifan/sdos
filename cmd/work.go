package cmd

import (
	"github.com/spf13/cobra"
	"os"

	"github.com/kuaifan/sdos/install"
	"github.com/wonderivan/logger"
)

// workCmd represents the websocket command
var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Simplest websocket",
	Long:  `sdos work --server-url ws://127.0.0.1`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if install.ServerUrl == "" {
			logger.Error("server-url is required.")
			err := cmd.Help()
			if err != nil {
				return
			}
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		install.BuildWork()
	},
}

func init() {
	rootCmd.AddCommand(workCmd)
	workCmd.Flags().StringVar(&install.ServerUrl, "server-url", "", "Websocket server url, \"ws://\" or \"wss://\" prefix.")
}
