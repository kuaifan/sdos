package cmd

import (
	"github.com/spf13/cobra"
	"os"

	"github.com/kuaifan/sdos/install"
)

var removeScriptCmd = &cobra.Command{
	Use:   "remove-script",
	Short: "Remove Script",
	PreRun: func(cmd *cobra.Command, args []string) {
		if install.ServerToken == "" {
			install.PrintError("The token is required.")
			os.Exit(0)
		}
		if install.ReportUrl == "" {
			install.PrintError("The report-url is required.")
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		install.RemoveScript()
	},
}

func init() {
	rootCmd.AddCommand(removeScriptCmd)
	removeScriptCmd.Flags().StringVar(&install.ServerToken, "token", "", "Token")
	removeScriptCmd.Flags().StringVar(&install.ReportUrl, "report-url", "", "Report url, \"http://\" or \"https://\" prefix.")
}
