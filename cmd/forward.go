package cmd

import (
	"github.com/kuaifan/sdos/install"
	"github.com/spf13/cobra"
	"os"
)

// forwardCmd represents the forward command
var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Only forward",
	PreRun: func(cmd *cobra.Command, args []string) {
		if install.ForwardConfig.Type != "add" && install.ForwardConfig.Type != "del" {
			err := cmd.Help()
			if err != nil {
				return
			}
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		install.BuildForward()
	},
}

func init() {
	rootCmd.AddCommand(forwardCmd)
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Type, "type", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Sport, "sport", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Eip, "eip", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Eport, "eport", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Protocol, "protocol", "", "")
}
