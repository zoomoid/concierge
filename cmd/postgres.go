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
	Script         string
	ContainerImage string
	JobName        string
	Timestamp      string
	Initiator      string
}

var scriptTemplate string = `
  psql postgresql://{{ .Settings.Connection.Username }}:{{ .Settings.Connection.Password }}@{{ .Settings.Connection.Host }}:{{ .Settings.Connection.Port }} <<EOF
  {{ if not .Options.NoUser }}
  CREATE USER {{ if .Options.UserIfNotExists }} \
    IF NOT EXISTS {{ end }} {{ .Settings.Username }} \
    WITH PASSWORD '{{ .Settings.Password }}';
  {{ end }}
  CREATE DATABASE {{ if .Options.DatabaseIfNotExists }} \
    IF NOT EXISTS {{ end }} {{ .Settings.Database }} \
    {{ if not .Options.NoUser }} WITH OWNER {{ end }} \
    ENCODING '{{ .Options.Encoding }}' \
    LC_COLLATE = '{{ .Options.Collation }}' \
    LC_CTYPE = '{{ .Options.ComparisonType }}';
	EOF
`

var jobSpec string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .JobName }}
  annotations:
    timestamp: {{ .Timestamp }}
    initator: {{ .Initiator }}
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
          - {{ .Script }}
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
	Database            string
	DatabaseVersion     string
	Username            string
	Password            string
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

			config := PostgresConfiguration{
				Options: &PostgresOptions{
					Encoding:            Encoding,
					Collation:           Collation,
					ComparisonType:      ComparisonType,
					UserIfNotExists:     UserIfNotExists,
					DatabaseIfNotExists: DatabaseIfNotExists,
					NoUser:              NoUser,
				},
				Settings: &PostgresSettings{
					Database: Database,
					Username: Username,
					Password: Password,
				},
				JobName:        fmt.Sprintf("concierge-%s-create-%s", "postgres", Database),
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
	createCmd.Flags().StringVarP(&Database, "database", "d", "", "Database name to create")
	createCmd.Flags().StringVarP(&Username, "username", "u", "", "Username of user to create")
	createCmd.Flags().StringVarP(&Password, "password", "p", "", "Password of user to create")
	createCmd.Flags().StringVarP(&Collation, "encoding", "", "UTF8", "Database encoding, psql ENCODING")
	createCmd.Flags().StringVarP(&Collation, "collation", "", "en_US.UTF-8", "Database collation, psql LC_COLLATE")
	createCmd.Flags().StringVarP(&ComparisonType, "comparsion", "lc-type", "en_US.UTF-8", "Database comparison type, psql LC_TYPE flag")
	createCmd.Flags().BoolVarP(&UserIfNotExists, "user.if-not-exists", "", false, "Query predicate IF NOT EXISTS for CREATE USER query")
	createCmd.Flags().BoolVarP(&DatabaseIfNotExists, "database.if-not-exists", "", false, "Query predicate IF NOT EXISTS for CREATE DATABASE query")
	createCmd.Flags().StringVarP(&DatabaseVersion, "database-version", "", "latest", "Database version to interact with. Keep this matching with your upstream database")
	createCmd.Flags().BoolVarP(&NoUser, "no-user", "", false, "Skips user creation and only creates a database")
}
