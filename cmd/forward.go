package cmd

import (
	"github.com/kuaifan/sdos/install"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// forwardCmd represents the forward command
var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Only forward",
	PreRun: func(cmd *cobra.Command, args []string) {
		install.FirewallConfig.Mode = strings.ToLower(install.FirewallConfig.Mode)
		install.FirewallConfig.Protocol = strings.ToLower(install.FirewallConfig.Protocol)
		if install.ForwardConfig.Mode != "add" && install.ForwardConfig.Mode != "del" {
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
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Mode, "mode", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Sport, "sport", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Eip, "eip", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Eport, "eport", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Protocol, "protocol", "", "")
}
