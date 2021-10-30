package cmd

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

type PostgresConfiguration struct {
	Options        *PostgresOptions
	Settings       *PostgresSettings
	Connection     *DatabaseConnection
	Script         string
	ContainerImage string
	JobName        string
	Timestamp      string
	Initiator      string
	Envs           map[string]string
}

var scriptTemplate string = `
		psql {{ if .Connection.Url }} {{ .Connection.Url }} {{ else }} postgresql://{{ .Connection.Username }}:{{ .Connection.Password }}@{{ .Connection.Host }}:{{ .Connection.Port }} {{ end }}<<EOF
		{{- if not .Options.NoUser }}
		CREATE USER {{ if .Options.UserIfNotExists }}IF NOT EXISTS{{ end }}{{ .Settings.Username }} WITH PASSWORD '{{ .Settings.Password }}';
		{{- end }}
		CREATE DATABASE{{ if .Options.DatabaseIfNotExists }} IF NOT EXISTS {{ end }}{{ .Settings.Database }}{{ if not .Options.NoUser }} WITH OWNER {{ end }}ENCODING '{{ .Options.Encoding }}'LC_COLLATE = '{{ .Options.Collation }}' LC_CTYPE = '{{ .Options.ComparisonType }}';
		EOF
`

var jobSpec string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .JobName }}
  annotations:
    timestamp: "{{ .Timestamp }}"
    initator: "{{ .Initiator }}"
spec:
  ttlSecondsAfterFinished: 100
  template:
    spec:
      containers:
      - name: creator
        image: {{ .ContainerImage }}
        command: 
          - /bin/sh
          - -c 
          - | {{ .Script }}
        env:
          {{ range $key, $value := .Envs }}
          - key: {{ $key }}
            value: {{ $value }}
          {{ end }}
      restartPolicy: Never
`

type PostgresOptions struct {
	// ENCODING flag for psql
	Encoding string
	// LC_COLLATION flag for psql
	Collation string
	// LC_CTYPE flag for psql
	ComparisonType string
	// IF NOT EXISTS for CREATE USER
	UserIfNotExists bool
	// IF NOT EXISTS for CREATE DATABASE
	DatabaseIfNotExists bool
	// Skip user creation and directly create database
	NoUser bool
}

type PostgresSettings struct {
	Database string
	Username string
	Password string
	Host     string
	Port     uint16
}

var (
	PostgresDatabase    string
	PostgresUsername    string
	PostgresPassword    string
	DatabaseVersion     string
	Encoding            string
	Collation           string
	ComparisonType      string
	UserIfNotExists     bool
	DatabaseIfNotExists bool
	NoUser              bool
	postgresCmd         = &cobra.Command{
		Use:   "postgres",
		Short: "Creates a new postgres database",
		Long:  "Templates a kubernetes job that creates a new postgres database",
		RunE: func(cmd *cobra.Command, args []string) error {
			var connection DatabaseConnection
			if len(Url) > 0 {
				connection = DatabaseConnection{
					Url: Url,
				}
			} else {
				connection = DatabaseConnection{
					Username: ConnectionUsername,
					Password: ConnectionPassword,
					Host:     Host,
					Port:     Port,
				}
			}

			config := PostgresConfiguration{
				Connection: &connection,
				Options: &PostgresOptions{
					Encoding:            Encoding,
					Collation:           Collation,
					ComparisonType:      ComparisonType,
					UserIfNotExists:     UserIfNotExists,
					DatabaseIfNotExists: DatabaseIfNotExists,
					NoUser:              NoUser,
				},
				Settings: &PostgresSettings{
					Database: PostgresDatabase,
					Username: PostgresUsername,
					Password: PostgresPassword,
				},
				JobName:        fmt.Sprintf("concierge-%s-create-%s", "postgres", PostgresDatabase),
				ContainerImage: fmt.Sprintf("postgres:%s", DatabaseVersion),
				Timestamp:      time.Now().UTC().String(),
				Script:         "",
				Initiator:      "user", // TODO(zoomoid): add kubeconfig user from $KUBECONFIG
			}
			tmpl, err := template.New("script").Parse(scriptTemplate)
			if err != nil {
				return err
			}
			b := &bytes.Buffer{}
			err = tmpl.Execute(b, config)
			if err != nil {
				return err
			}
			config.Script = b.String()

			tmpl, err = template.New("jobSpec").Parse(jobSpec)
			if err != nil {
				return err
			}
			err = tmpl.Execute(os.Stdout, config)
			if err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	createCmd.AddCommand(postgresCmd)
	postgresCmd.Flags().StringVarP(&PostgresDatabase, "database", "d", "", "Database name to create")
	postgresCmd.Flags().StringVarP(&PostgresUsername, "username", "u", "", "Username of user to create")
	postgresCmd.Flags().StringVarP(&PostgresPassword, "password", "p", "", "Password of user to create")
	postgresCmd.Flags().StringVar(&Collation, "encoding", "UTF8", "Database encoding, psql ENCODING")
	postgresCmd.Flags().StringVar(&Collation, "collation", "en_US.UTF-8", "Database collation, psql LC_COLLATE")
	postgresCmd.Flags().StringVar(&ComparisonType, "comparison", "en_US.UTF-8", "Database comparison type, psql LC_TYPE flag")
	postgresCmd.Flags().BoolVar(&UserIfNotExists, "user-if-not-exists", false, "Query predicate IF NOT EXISTS for CREATE USER query")
	postgresCmd.Flags().BoolVar(&DatabaseIfNotExists, "database-if-not-exists", false, "Query predicate IF NOT EXISTS for CREATE DATABASE query")
	postgresCmd.Flags().StringVar(&DatabaseVersion, "database-version", "latest", "Database version to interact with. Keep this matching with your upstream database")
	postgresCmd.Flags().BoolVar(&NoUser, "no-user", false, "Skips user creation and only creates a database")
}
