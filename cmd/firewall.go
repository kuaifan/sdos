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
	Short: "Only firewall",
	PreRun: func(cmd *cobra.Command, args []string) {
		install.FirewallConfig.Mode = strings.ToLower(install.FirewallConfig.Mode)
		install.FirewallConfig.Protocol = strings.ToLower(install.FirewallConfig.Protocol)
		if !install.InArray(install.FirewallConfig.Mode, []string{"add", "del", "install", "uninstall"}) {
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
	firewallCmd.Flags().StringVar(&install.FirewallConfig.Mode, "mode", "", "add|del|install|uninstall")
	firewallCmd.Flags().StringVar(&install.FirewallConfig.Ports, "ports", "", "")
	firewallCmd.Flags().StringVar(&install.FirewallConfig.Type, "type", "", "")
	firewallCmd.Flags().StringVar(&install.FirewallConfig.Address, "address", "", "")
	firewallCmd.Flags().StringVar(&install.FirewallConfig.Protocol, "protocol", "", "")
	firewallCmd.Flags().BoolVar(&install.FirewallConfig.Force, "force", false, "delete before adding")
}
