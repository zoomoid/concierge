package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type DatabaseConnection struct {
	Url      string
	Host     string
	Port     string
	Username string
	Password string
}

var (
	Url                string
	Host               string
	Port               string
	ConnectionUsername string
	ConnectionPassword string
	createCmd          = &cobra.Command{
		Use:   "create",
		Short: "Creates a new database",
		Long:  `Templates a new database from a given type, name, user, and password`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("missing database type")
		},
	}
)

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&Url, "url", "C", "", "A fully-qualified connection URL to the database")
	createCmd.Flags().StringVarP(&Host, "host", "H", "", "The database hostname to connect to")
	createCmd.Flags().StringVarP(&Port, "port", "P", "", "The database port at which to connect to")
	createCmd.Flags().StringVarP(&ConnectionUsername, "username", "U", "", "Database user to connect with")
	createCmd.Flags().StringVarP(&ConnectionPassword, "password", "A", "", "Database password to connect with")
}
