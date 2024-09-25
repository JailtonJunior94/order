package main

import (
	"context"
	"log"

	"github.com/jailtonjunior94/outbox/cmd/consumer"
	"github.com/jailtonjunior94/outbox/cmd/server"
	"github.com/jailtonjunior94/outbox/cmd/worker"
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

	server := &cobra.Command{
		Use:   "api",
		Short: "Outbox API",
		Run: func(cmd *cobra.Command, args []string) {
			server.NewApiServer().Run()
		},
	}

	consumers := &cobra.Command{
		Use:   "consumers",
		Short: "Outbox Consumers",
		Run: func(cmd *cobra.Command, args []string) {
			consumer.NewConsumer().Run()
		},
	}

	workers := &cobra.Command{
		Use:   "workers",
		Short: "Outbox Workers",
		Run: func(cmd *cobra.Command, args []string) {
			worker.NewWorkers().Run()
		},
	}

	root.AddCommand(migrate, server, consumers, workers)
	root.Execute()
}
