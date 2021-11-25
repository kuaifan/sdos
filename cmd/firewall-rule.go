package cmd

import (
	"github.com/kuaifan/sdos/install"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// firewallRuleCmd represents the firewallRule command
var firewallRuleCmd = &cobra.Command{
	Use:   "firewall-rule",
	Short: "Firewall rule",
	PreRun: func(cmd *cobra.Command, args []string) {
		install.FirewallRuleConfig.Mode = strings.ToLower(install.FirewallRuleConfig.Mode)
		install.FirewallRuleConfig.Protocol = strings.ToLower(install.FirewallRuleConfig.Protocol)
		if !install.InArray(install.FirewallRuleConfig.Mode, []string{"add", "del"}) {
			install.Error("mode error")
			os.Exit(0)
		}
		if install.FirewallRuleConfig.Key == "" {
			install.Error("key error")
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		install.BuildFirewallRule()
	},
}

func init() {
	rootCmd.AddCommand(firewallRuleCmd)
	firewallRuleCmd.Flags().StringVar(&install.FirewallRuleConfig.Mode, "mode", "", "add or del")
	firewallRuleCmd.Flags().StringVar(&install.FirewallRuleConfig.Ports, "ports", "", "")
	firewallRuleCmd.Flags().StringVar(&install.FirewallRuleConfig.Type, "type", "", "")
	firewallRuleCmd.Flags().StringVar(&install.FirewallRuleConfig.Address, "address", "", "")
	firewallRuleCmd.Flags().StringVar(&install.FirewallRuleConfig.Protocol, "protocol", "", "")
	firewallRuleCmd.Flags().StringVar(&install.FirewallRuleConfig.Key, "key", "", "")
}
