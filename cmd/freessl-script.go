package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"

	"github.com/kuaifan/sdos/install"
)

var freeSSlScriptCmd = &cobra.Command{
	Use:   "freessl-script",
	Short: "FreeSSL Script",
	PreRun: func(cmd *cobra.Command, args []string) {

		if install.ReportUrl == "" || install.ServerDomain == "" {
			install.PrintError("The report-url / server-domain are required.")
			os.Exit(0)
		}

		ip := install.GetIp()

		ips, _ := net.LookupHost(install.ServerDomain)
		if install.StringsContains(ips, ip) == -1 {
			install.PrintError(fmt.Sprintf("Domain name [%s] resolution results %s, inconsistent with server IP %s.", install.ServerDomain, ips, ip))
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		install.FreeSSLNode()
	},
}

func init() {
	rootCmd.AddCommand(freeSSlScriptCmd)
	freeSSlScriptCmd.Flags().StringVar(&install.ServerDomain, "server-domain", "", "Server domain, example: w1.abc.com")
	freeSSlScriptCmd.Flags().StringVar(&install.ReportUrl, "report-url", "", "Report url, \"http://\" or \"https://\" prefix.")
}
