package cmd

import "github.com/spf13/cobra"

var (
	MysqlDatabase string
	MysqlPassword string
	MysqlUsername string
	mysqlCmd      = &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(mysqlCmd)

	mysqlCmd.Flags().StringVarP(&MysqlDatabase, "database", "d", "", "Database name to create")
	mysqlCmd.Flags().StringVarP(&MysqlUsername, "username", "u", "", "Username of user to create")
	mysqlCmd.Flags().StringVarP(&MysqlPassword, "password", "p", "", "Password of user to create")
}
