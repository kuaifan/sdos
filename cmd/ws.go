package cmd

import (
	"github.com/spf13/cobra"
	"os"

	"github.com/kuaifan/sdos/install"
	"github.com/wonderivan/logger"
)

// wsCmd represents the websocket command
var wsCmd = &cobra.Command{
	Use:   "ws",
	Short: "Simplest websocket",
	Long:  `sdos ws --server-url ws://127.0.0.1`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(install.ServerUrl) == 0 {
			logger.Error("server-url is empty at the same time. please check your args in command.")
			err := cmd.Help()
			if err != nil {
				return
			}
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		install.BuildWs()
	},
}

func init() {
	rootCmd.AddCommand(wsCmd)
	wsCmd.Flags().StringVar(&install.ServerUrl, "server-url", "", "websocket url")
}
