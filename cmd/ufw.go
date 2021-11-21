package cmd

import (
	"github.com/kuaifan/sdos/install"
	"github.com/spf13/cobra"
	"os"
)

// ufwCmd represents the ufw command
var ufwCmd = &cobra.Command{
	Use:   "ufw",
	Short: "Only ufw",
	PreRun: func(cmd *cobra.Command, args []string) {
		if install.UFWConfig.Type != "add" && install.UFWConfig.Type != "del" {
			err := cmd.Help()
			if err != nil {
				return
			}
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		install.BuildUfw()
	},
}

func init() {
	rootCmd.AddCommand(ufwCmd)
	ufwCmd.Flags().StringVar(&install.UFWConfig.Type, "type", "", "")
	ufwCmd.Flags().StringVar(&install.UFWConfig.Sport, "sport", "", "")
	ufwCmd.Flags().StringVar(&install.UFWConfig.Dport, "dport", "", "")
	ufwCmd.Flags().StringVar(&install.UFWConfig.Dip, "dip", "", "")
	ufwCmd.Flags().StringVar(&install.UFWConfig.Protocol, "protocol", "", "")
	ufwCmd.Flags().StringVar(&install.UFWConfig.Path, "path", "/etc/ufw/before.rules", "")
}
