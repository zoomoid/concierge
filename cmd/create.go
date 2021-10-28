package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createCmd)
}

var (
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Creates a new database",
		Long:  `Templates a new database from a given type, name, user, and password`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("missing database type")
		},
	}
)
