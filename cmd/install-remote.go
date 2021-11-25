package cmd

import (
	"github.com/spf13/cobra"
	"net"
	"os"

	"github.com/kuaifan/sdos/install"
)

// installRemoteCmd represents the installRemote command
var installRemoteCmd = &cobra.Command{
	Use:   "install-remote",
	Short: "Install Remote",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(install.NodeIPs) == 0 {
			ipv4, _, _ := install.RunCommand("-c", "curl -4 ip.sb")
			address := net.ParseIP(ipv4)
			if address != nil {
				install.NodeIPs = append(install.NodeIPs, ipv4)
			}
		}
		if len(install.NodeIPs) == 0 || install.ReportUrl == "" {
			install.RemoteError("node / report-url required.")
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
		install.BuildInstallRemote(beforeNodes)
	},
}

func init() {
	rootCmd.AddCommand(installRemoteCmd)
	installRemoteCmd.Flags().StringSliceVar(&install.NodeIPs, "node", []string{}, "Multi nodes ex. 192.168.0.5-192.168.0.5")
	installRemoteCmd.Flags().StringVar(&install.SSHConfig.User, "user", "root", "Servers user name for ssh")
	installRemoteCmd.Flags().StringVar(&install.SSHConfig.Password, "passwd", "", "Password for ssh, It’s base64 encode")
	installRemoteCmd.Flags().StringVar(&install.ReportUrl, "report-url", "", "Report url, \"http://\" or \"https://\" prefix.")
}
