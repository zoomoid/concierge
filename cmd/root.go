package cmd

import (
	"github.com/spf13/cobra"
)

var (
	DryRun  bool
	Verbose bool
	rootCmd = &cobra.Command{
		Use:   "concierge",
		Short: "Bootstraps your database from inside a Kubernetes cluster that has access to the DB",
		Long:  "Concierge is a minimal Golang CLI made for interacting with kubectl to draft Job specs for creating and bootstrapping databases on postgres and mysql",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {

}
