package cli

import (
	"github.com/mustafmst/taskerr/internal/data"
	"github.com/spf13/cobra"
)

func SetupCLIApp(dataService *data.Service) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "taskerr",
		Short: "Taskerr CLI Application",
	}

	rootCmd.AddCommand(
		newListCommand(dataService),
		newAddCommand(dataService),
		newTagCommand(dataService),
		newTaskCommand(dataService),
	)

	return rootCmd
}
