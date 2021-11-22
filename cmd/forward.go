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
		install.ForwardConfig.Mode = strings.ToLower(install.ForwardConfig.Mode)
		install.ForwardConfig.Protocol = strings.ToLower(install.ForwardConfig.Protocol)
		if !install.InArray(install.ForwardConfig.Mode, []string{"add", "del"}) {
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
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Mode, "mode", "", "add or del")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Sport, "sport", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Dip, "dip", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Dport, "dport", "", "")
	forwardCmd.Flags().StringVar(&install.ForwardConfig.Protocol, "protocol", "", "")
}
