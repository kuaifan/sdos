package cmd

import (
	"github.com/spf13/cobra"
	"net"
	"os"

	"github.com/kuaifan/sdos/install"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Exec",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(install.NodeIPs) == 0 {
			ipv4, _, _ := install.RunCommand("-c", "curl -4 ip.sb")
			address := net.ParseIP(ipv4)
			if address != nil {
				install.NodeIPs = append(install.NodeIPs, ipv4)
			}
		}
		if len(install.NodeIPs) == 0 {
			install.RemoteError("node required.")
			err := cmd.Help()
			if err != nil {
				return
			}
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		beforeNodes := install.ParseIPs(install.NodeIPs)
		install.BuildExec(beforeNodes)
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.Flags().StringSliceVar(&install.NodeIPs, "node", []string{}, "Multi nodes ex. 192.168.0.5-192.168.0.5")
	installRemoteCmd.Flags().StringVar(&install.SSHConfig.User, "user", "root", "Servers user name for ssh")
	installRemoteCmd.Flags().StringVar(&install.SSHConfig.Password, "passwd", "", "Password for ssh, It’s base64 encode")
	execCmd.Flags().StringVar(&install.ExecConfig.Cmd, "cmd", "", "")
	execCmd.Flags().StringVar(&install.ExecConfig.Param, "param", "", "")
}
