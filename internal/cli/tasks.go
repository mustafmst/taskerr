package cli

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/mustafmst/taskerr/internal/data"
	"github.com/spf13/cobra"
)

func newListCommand(dataService *data.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List all tasks",
		Run: func(cmd *cobra.Command, args []string) {
			tasksList, err := dataService.TasksRepo.GetAllWithTags()
			if err != nil {
				log.Fatalf("Error retrieving tasks: %v", err)
			}

			for _, task := range tasksList {
				tagNames := make([]string, len(task.Tags))
				for i, tag := range task.Tags {
					tagNames[i] = tag.Name
				}

				tagsStr := ""
				if len(tagNames) > 0 {
					tagsStr = fmt.Sprintf(" [%s]", strings.Join(tagNames, ", "))
				}

				fmt.Printf("ID: %d, Title: %s, Completed: %t%s\n", task.ID, task.Description, task.State, tagsStr)
			}
		},
	}
}

func newAddCommand(dataService *data.Service) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add [task description]",
		Short: "Add a new task",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Please provide a task description")
				return
			}

			description := strings.Join(args, " ")
			details, _ := cmd.Flags().GetString("details")
			tagNames := parseCommaSeparated(cmd, "tags")
			task, err := dataService.CreateTask(description, details, nil, tagNames)
			if err != nil {
				log.Fatalf("Error creating task: %v", err)
			}

			fmt.Printf("Task created with ID: %d\n", task.ID)
		},
	}

	addCmd.Flags().String("details", "", "Longer task description")
	addCmd.Flags().String("tags", "", "Comma-separated list of tags (e.g., --tags=work,urgent)")
	return addCmd
}

func newTaskCommand(dataService *data.Service) *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}

	taskCmd.AddCommand(
		newTaskTagCommand(dataService),
		newTaskUntagCommand(dataService),
	)

	return taskCmd
}

func newTaskTagCommand(dataService *data.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "tag [task_id] [tag_name]",
		Short: "Attach a tag to a task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			taskID := parseTaskID(args[0])
			tag, err := dataService.AttachTagToTask(taskID, args[1])
			if err != nil {
				log.Fatalf("Error attaching tag to task: %v", err)
			}

			fmt.Printf("Tag '%s' attached to task %d\n", tag.Name, taskID)
		},
	}
}

func newTaskUntagCommand(dataService *data.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "untag [task_id] [tag_name]",
		Short: "Remove a tag from a task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			taskID := parseTaskID(args[0])
			if err := dataService.DetachTagFromTask(taskID, args[1]); err != nil {
				log.Fatalf("Error removing tag from task: %v", err)
			}

			fmt.Printf("Tag '%s' removed from task %d\n", args[1], taskID)
		},
	}
}

func parseTaskID(raw string) uint {
	taskID, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		log.Fatalf("Invalid task ID: %v", err)
	}

	return uint(taskID)
}

func parseCommaSeparated(cmd *cobra.Command, flagName string) []string {
	value, _ := cmd.Flags().GetString(flagName)
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
