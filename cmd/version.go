package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of concierge",
	Long:  `All software has versions. This is concierge's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("concierge v0.0.1-alpha.1 -- HEAD")
	},
}
