package cmd

import (
	"github.com/spf13/cobra"
	"net"
	"os"

	"github.com/kuaifan/sdos/install"
	"github.com/kuaifan/sdos/pkg/logger"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(install.NodeIPs) == 0 {
			ipv4, _, _ := install.RunCommand("-c", "curl -4 ip.sb")
			address := net.ParseIP(ipv4)
			if address != nil {
				install.NodeIPs = append(install.NodeIPs, ipv4)
			}
		}
		if len(install.NodeIPs) == 0 || install.ReportUrl == "" {
			logger.Error("node / report-url required.")
			err := cmd.Help()
			if err != nil {
				return
			}
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		beforeNodes := install.ParseIPs(install.NodeIPs)
		if install.SSHConfig.User == "" {
			install.SSHConfig.User = "root"
		}
		if install.SSHConfig.Password != "" {
			install.SSHConfig.Password = install.Base64Decode(install.SSHConfig.Password)
		}
		install.BuildRemove(beforeNodes)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().StringSliceVar(&install.NodeIPs, "node", []string{}, "Multi nodes ex. 192.168.0.5-192.168.0.5")
	removeCmd.Flags().StringVar(&install.SSHConfig.User, "user", "root", "Servers user name for ssh")
	removeCmd.Flags().StringVar(&install.SSHConfig.Password, "passwd", "", "Password for ssh")
	removeCmd.Flags().StringVar(&install.ReportUrl, "report-url", "", "Report url, \"http://\" or \"https://\" prefix.")
}
