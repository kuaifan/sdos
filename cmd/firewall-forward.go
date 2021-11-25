package cmd

import (
	"github.com/kuaifan/sdos/install"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// firewallForwardCmd represents the firewallForward command
var firewallForwardCmd = &cobra.Command{
	Use:   "firewall-forward",
	Short: "Firewall forward",
	PreRun: func(cmd *cobra.Command, args []string) {
		install.FirewallForwardConfig.Mode = strings.ToLower(install.FirewallForwardConfig.Mode)
		install.FirewallForwardConfig.Protocol = strings.ToLower(install.FirewallForwardConfig.Protocol)
		if !install.InArray(install.FirewallForwardConfig.Mode, []string{"add", "del"}) {
			install.Error("mode error")
			os.Exit(0)
		}
		if install.FirewallForwardConfig.Key == "" {
			install.Error("key error")
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		install.BuildFirewallForward()
	},
}

func init() {
	rootCmd.AddCommand(firewallForwardCmd)
	firewallForwardCmd.Flags().StringVar(&install.FirewallForwardConfig.Mode, "mode", "", "add or del")
	firewallForwardCmd.Flags().StringVar(&install.FirewallForwardConfig.Sport, "sport", "", "")
	firewallForwardCmd.Flags().StringVar(&install.FirewallForwardConfig.Dip, "dip", "", "")
	firewallForwardCmd.Flags().StringVar(&install.FirewallForwardConfig.Dport, "dport", "", "")
	firewallForwardCmd.Flags().StringVar(&install.FirewallForwardConfig.Protocol, "protocol", "", "")
	firewallForwardCmd.Flags().StringVar(&install.FirewallForwardConfig.Key, "key", "", "")
}
