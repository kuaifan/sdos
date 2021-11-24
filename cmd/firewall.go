package cmd

import (
	"github.com/kuaifan/sdos/install"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// firewallCmd represents the firewall command
var firewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Firewall",
	PreRun: func(cmd *cobra.Command, args []string) {
		install.FirewallConfig.Mode = strings.ToLower(install.FirewallConfig.Mode)
		if !install.InArray(install.FirewallConfig.Mode, []string{"install", "uninstall"}) {
			err := cmd.Help()
			if err != nil {
				return
			}
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		install.BuildFirewall()
	},
}

func init() {
	rootCmd.AddCommand(firewallCmd)
	firewallCmd.Flags().StringVar(&install.FirewallConfig.Mode, "mode", "", "install|uninstall")
}
