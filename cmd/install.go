package cmd

import (
	"github.com/spf13/cobra"
	"net"
	"os"

	"github.com/kuaifan/sdos/install"
	"github.com/kuaifan/sdos/pkg/logger"
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
			logger.Error("node / manage-image / server-url required.")
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
		install.ServerToken = install.RandomString(32)
		install.BuildInstall(beforeNodes)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringSliceVar(&install.NodeIPs, "node", []string{}, "Multi nodes ex. 192.168.0.5-192.168.0.5")
	installCmd.Flags().StringVar(&install.SSHConfig.User, "user", "root", "Servers user name for ssh")
	installCmd.Flags().StringVar(&install.SSHConfig.Password, "passwd", "", "Password for ssh")
	installCmd.Flags().BoolVar(&install.InstallReset, "reset", false, "Remove before installation")
	installCmd.Flags().StringVar(&install.ManageImage, "manage-image", "", "Image of Management")
	installCmd.Flags().StringVar(&install.ServerUrl, "server-url", "", "Release server url, \"http://\" or \"https://\" prefix.")
}
