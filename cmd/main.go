package main

import (
	"context"
	"log"

	"github.com/jailtonjunior94/outbox/pkg/bundle"
	migration "github.com/jailtonjunior94/outbox/pkg/database/migrate"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "outbox",
		Short: "Outbox",
	}

	migrate := &cobra.Command{
		Use:   "migrate",
		Short: "Outbox Migrations",
		Run: func(cmd *cobra.Command, args []string) {
			container := bundle.NewContainer(context.Background())
			migrate, err := migration.NewMigrateCockroachDB(container.DB, container.Config.DBConfig.MigratePath, container.Config.DBConfig.Name)
			if err != nil {
				log.Fatal(err)
			}
			if err = migrate.Execute(); err != nil {
				log.Fatal(err)
			}
		},
	}

	api := &cobra.Command{
		Use:   "api",
		Short: "Outbox API",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("not implement")
		},
	}

	consumers := &cobra.Command{
		Use:   "consumers",
		Short: "Outbox Consumers",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("not implement")
		},
	}

	workers := &cobra.Command{
		Use:   "workers",
		Short: "Outbox Workers",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("not implement")
		},
	}

	root.AddCommand(migrate, api, consumers, workers)
	root.Execute()
}
