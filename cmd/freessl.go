package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net"
	"os"

	"github.com/kuaifan/sdos/install"
	"github.com/kuaifan/sdos/pkg/logger"
)

// freesslCmd represents the freessl command
var freesslCmd = &cobra.Command{
	Use:   "freessl",
	Short: "Freessl",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(install.NodeIPs) == 0 {
			ipv4, _, _ := install.RunCommand("-c", "curl -4 ip.sb")
			address := net.ParseIP(ipv4)
			if address != nil {
				install.NodeIPs = append(install.NodeIPs, ipv4)
			}
		}
		if len(install.NodeIPs) == 0 || install.ReportUrl == "" || install.ServerDomain == "" {
			logger.Error("node / report-url / server-domain required.")
			err := cmd.Help()
			if err != nil {
				return
			}
			os.Exit(0)
		}
		if len(install.NodeIPs) > 1 {
			install.Error("Only one host is supported when filling in the domain name.")
			os.Exit(0)
		}
		ip, _ := net.LookupHost(install.ServerDomain)
		if install.StringsContains(ip, install.NodeIPs[0]) == -1 {
			install.Error(fmt.Sprintf("Domain name [%s] resolution results %s, inconsistent with server IP %s.", install.ServerDomain, ip, install.NodeIPs[0]))
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
		install.BuildFreessl(beforeNodes)
	},
}

func init() {
	rootCmd.AddCommand(freesslCmd)
	freesslCmd.Flags().StringSliceVar(&install.NodeIPs, "node", []string{}, "Multi nodes ex. 192.168.0.5-192.168.0.5")
	freesslCmd.Flags().StringVar(&install.SSHConfig.User, "user", "root", "Servers user name for ssh")
	freesslCmd.Flags().StringVar(&install.SSHConfig.Password, "passwd", "", "Password for ssh")
	freesslCmd.Flags().StringVar(&install.ServerDomain, "server-domain", "", "Server domain, example: w1.abc.com")
	freesslCmd.Flags().StringVar(&install.ReportUrl, "report-url", "", "Report url, \"http://\" or \"https://\" prefix.")
}
