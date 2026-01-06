package cli

import "github.com/spf13/cobra"

func SetupCliApp() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "taskerr",
		Short: "Taskerr CLI Application",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all tasks",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Listing all tasks...")
		},
	}
	rootCmd.AddCommand(lsCmd)
	return rootCmd
}
