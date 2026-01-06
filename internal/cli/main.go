package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/mustafmst/taskerr/internal/data"
	"github.com/mustafmst/taskerr/internal/data/tasks"
	"github.com/spf13/cobra"
)

func SetupCliApp(dataService *data.Service) *cobra.Command {
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
			tasks, err := dataService.TasksRepo.GetAll()
			if err != nil {
				log.Fatalf("Error retrieving tasks: %v", err)
			}
			for _, task := range tasks {
				fmt.Printf("ID: %d, Title: %s, Completed: %t\n", task.ID, task.Description, task.State)
			}
		},
	}
	addCmd := &cobra.Command{
		Use:   "add [task description]",
		Short: "Add a new task",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Adding task: %s\n", strings.Join(args, " "))
			dataService.TasksRepo.Create(&tasks.Task{Description: strings.Join(args, " "), State: false})
		},
	}
	rootCmd.AddCommand(lsCmd, addCmd)
	return rootCmd
}
