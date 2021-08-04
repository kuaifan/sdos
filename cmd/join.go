package cmd

import (
	"github.com/spf13/cobra"
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
			logger.Error("this command is join feature, node is empty at the same time. please check your args in command.")
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
	joinCmd.Flags().StringSliceVar(&install.NodeIPs, "node", []string{}, "sdwan multi-nodes ex. 192.168.0.5-192.168.0.5")
	joinCmd.Flags().StringVar(&install.SSHConfig.User, "user", "root", "servers user name for ssh")
	joinCmd.Flags().StringVar(&install.SSHConfig.Password, "passwd", "", "password for ssh")
	joinCmd.Flags().StringVar(&install.ServerUrl, "server-url", "", "node publishes to this URL after deployment is complete")
}
