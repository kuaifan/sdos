package cmd

import (
	"github.com/kuaifan/sdos/install"
	"github.com/spf13/cobra"
)

// netCmd
var netCmd = &cobra.Command{
	Use:   "net",
	Short: "Get net information",
	Run: func(cmd *cobra.Command, args []string) {
		install.BuildNet()
	},
}

func init() {
	rootCmd.AddCommand(netCmd)
	netCmd.Flags().StringVarP(&install.NetInterface, "interface", "i", "*", "interface")
	netCmd.Flags().UintVarP(&install.NetCount, "count", "c", 1, "count")
	netCmd.Flags().Float64VarP(&install.NetUpdateTime, "update", "t", 2, "update time")
}
