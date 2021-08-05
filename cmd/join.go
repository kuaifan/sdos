package cmd

import (
	"github.com/spf13/cobra"
	"net"
	"os"

	"github.com/kuaifan/sdos/install"
	"github.com/wonderivan/logger"
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Simplest way to join your sdwan cluster",
	Long:  `sdos join --node 192.168.0.5`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(install.NodeIPs) == 0 {
			ipv4, _, _ := install.RunShellInSystem("curl -4 ip.sb")
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
		install.BuildJoin(beforeNodes)
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)
	joinCmd.Flags().StringSliceVar(&install.NodeIPs, "node", []string{}, "Multi nodes ex. 192.168.0.5-192.168.0.5")
	joinCmd.Flags().StringVar(&install.SSHConfig.User, "user", "root", "Servers user name for ssh")
	joinCmd.Flags().StringVar(&install.SSHConfig.Password, "passwd", "", "Password for ssh")
	joinCmd.Flags().StringVar(&install.ManageImage, "manage-image", "", "Image of Management")
	joinCmd.Flags().StringVar(&install.ServerUrl, "server-url", "", "Release server url, \"http://\" or \"https://\" prefix.")
}
