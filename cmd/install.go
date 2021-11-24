package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net"
	"os"

	"github.com/kuaifan/sdos/install"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Simplest way to install your sdwan cluster",
	Long:  `sdos install --node 192.168.0.5`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(install.NodeIPs) == 0 {
			ipv4, _, _ := install.RunCommand("-c", "curl -4 ip.sb")
			address := net.ParseIP(ipv4)
			if address != nil {
				install.NodeIPs = append(install.NodeIPs, ipv4)
			}
		}
		if len(install.NodeIPs) == 0 || install.ManageImage == "" || install.ServerUrl == "" {
			install.Error("node / manage-image / server-url required.")
			err := cmd.Help()
			if err != nil {
				return
			}
			os.Exit(0)
		}
		if install.ServerKey != "" {
			if install.ServerCrt == "" {
				install.Error("Key exist, crt required")
				os.Exit(0)
			}
			if install.ServerDomain == "" {
				install.Error("Key exist, domain required")
				os.Exit(0)
			}
		}
		if install.ServerCrt != "" {
			if install.ServerKey == "" {
				install.Error("Crt exist, key required")
				os.Exit(0)
			}
			if install.ServerDomain == "" {
				install.Error("Crt exist, domain required")
				os.Exit(0)
			}
		}
		if install.ServerDomain != "" && install.ServerKey == "" {
			if len(install.NodeIPs) > 1 {
				install.Error("Only one host is supported when filling in the domain name.")
				os.Exit(0)
			}
			ip, _ := net.LookupHost(install.ServerDomain)
			if install.StringsContains(ip, install.NodeIPs[0]) == -1 {
				install.Error(fmt.Sprintf("Domain name [%s] resolution results %s, inconsistent with server IP %s.", install.ServerDomain, ip, install.NodeIPs[0]))
				os.Exit(0)
			}
		}
		if install.ReportUrl == "" {
			install.ReportUrl = install.ServerUrl
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
		install.ServerToken = install.RandomString(32)
		install.BuildInstall(beforeNodes)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringSliceVar(&install.NodeIPs, "node", []string{}, "Multi nodes ex. 192.168.0.5-192.168.0.5")
	installCmd.Flags().StringVar(&install.SSHConfig.User, "user", "root", "Servers user name for ssh")
	installCmd.Flags().StringVar(&install.SSHConfig.Password, "passwd", "", "Password for ssh, It’s base64 encode")
	installCmd.Flags().StringVar(&install.Mtu, "mtu", "", "Maximum transmission unit")
	installCmd.Flags().StringVar(&install.ManageImage, "manage-image", "", "Image of management")
	installCmd.Flags().StringVar(&install.ServerUrl, "server-url", "", "Server url, \"http://\" or \"https://\" prefix.")
	installCmd.Flags().StringVar(&install.ServerDomain, "server-domain", "", "Server domain, example: w1.abc.com")
	installCmd.Flags().StringVar(&install.ServerKey, "server-key", "", "Server domain key")
	installCmd.Flags().StringVar(&install.ServerCrt, "server-crt", "", "Server domain certificate")
	installCmd.Flags().StringVar(&install.ReportUrl, "report-url", "", "Report url, \"http://\" or \"https://\" prefix, default to server-url.")
	installCmd.Flags().StringVar(&install.SwapFile, "swap", "", "Add swap partition, unit MB")
	installCmd.Flags().BoolVar(&install.InReset, "reset", false, "Remove before installation")
}
