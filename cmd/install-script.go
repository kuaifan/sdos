package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/kuaifan/sdos/install"
	"github.com/nahid/gohttp"
	"github.com/spf13/cobra"
)

var installScriptCmd = &cobra.Command{
	Use:   "install-script",
	Short: "Install without password",
	PreRun: func(cmd *cobra.Command, args []string) {

		if install.ServerToken == "" {
			install.PrintError("The token is required!")
			os.Exit(0)
		}

		if install.ManageImage == "" || install.ServerUrl == "" {
			install.PrintError("The manage-image and server-url are required!")
			os.Exit(0)
		}

		fmt.Print("Checking arguments...")

		if install.ServerDomain != "" {
			if install.CustomCrt {
				resp, err := gohttp.NewRequest().Get(fmt.Sprintf("%s/api/get-crt?token=%s", install.ServerUrl, install.ServerToken))
				if err != nil {
					install.PrintError(fmt.Sprintf("\rFailed to get custom certificate: %s\n", err.Error()))
					return
				}

				body, _ := resp.GetBodyAsByte()

				data := make(map[string]string)
				_ = json.Unmarshal(body, &data)

				install.ServerKey = data["key"]
				install.ServerCrt = data["crt"]

				if install.ServerKey == "" || install.ServerCrt == "" {
					install.PrintError(fmt.Sprintf("\rCertificate content format error: %s\n", err.Error()))
					os.Exit(0)
				}
			} else {
				ip := install.GetIp()
				ips, _ := net.LookupHost(install.ServerDomain)
				if install.StringsContains(ips, ip) == -1 {
					install.PrintError(fmt.Sprintf("\rDomain name [%s] resolution results %s, inconsistent with server IP %s.", install.ServerDomain, ips, ip))
					os.Exit(0)
				}
			}
		}

		if install.ReportUrl == "" {
			install.ReportUrl = install.ServerUrl
		}

		_, err := gohttp.NewRequest().Head(install.ServerUrl)
		if err != nil {
			install.PrintError(fmt.Sprintf("\rWrong server-url: %s\n", err.Error()))
			os.Exit(0)
		}

		_, err = gohttp.NewRequest().Head(install.ReportUrl)
		if err != nil {
			install.PrintError(fmt.Sprintf("\rWrong report-url: %s\n", err.Error()))
			os.Exit(0)
		}

	},
	Run: func(cmd *cobra.Command, args []string) {
		install.ScriptInstallNode()
	},
}

func init() {
	rootCmd.AddCommand(installScriptCmd)
	installScriptCmd.Flags().StringVar(&install.ServerToken, "token", "", "Token")
	installScriptCmd.Flags().StringVar(&install.Mtu, "mtu", "", "Maximum transmission unit")
	installScriptCmd.Flags().StringVar(&install.ManageImage, "manage-image", "", "Image of management")
	installScriptCmd.Flags().StringVar(&install.ServerUrl, "server-url", "", "Server url, \"http://\" or \"https://\" prefix.")
	installScriptCmd.Flags().StringVar(&install.ServerDomain, "server-domain", "", "Server domain, example: w1.abc.com")
	installScriptCmd.Flags().BoolVar(&install.CustomCrt, "custom-crt", false, "Custom certificate")
	installScriptCmd.Flags().StringVar(&install.ReportUrl, "report-url", "", "Report url, \"http://\" or \"https://\" prefix, default to server-url.")
	installScriptCmd.Flags().StringVar(&install.SwapFile, "swap", "", "Add swap partition, unit MB")
	installScriptCmd.Flags().BoolVar(&install.InReset, "reset", false, "Remove before installation")
}
