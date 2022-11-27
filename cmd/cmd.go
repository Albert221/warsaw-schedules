package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"warsaw-schedules.dev/db"
)

var rootCmd = &cobra.Command{}

func init() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(parseCmd)
}

type ctxKey int

const stopRepoKey ctxKey = iota

func Execute(stopRepo db.StopRepository) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, stopRepoKey, stopRepo)

	return rootCmd.ExecuteContext(ctx)
}
