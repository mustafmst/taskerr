package cli

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/mustafmst/taskerr/internal/data"
	"github.com/mustafmst/taskerr/internal/data/tasks"
	"github.com/spf13/cobra"
)

func SetupCLIApp(dataService *data.Service) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "taskerr",
		Short: "Taskerr CLI Application",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	// List tasks command
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all tasks",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Listing all tasks...")
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

	// Add task command with --tags flag
	addCmd := &cobra.Command{
		Use:   "add [task description]",
		Short: "Add a new task",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Please provide a task description")
				return
			}
			description := strings.Join(args, " ")
			fmt.Printf("Adding task: %s\n", description)

			task := &tasks.Task{Description: description, State: false}
			if err := dataService.TasksRepo.Create(task); err != nil {
				log.Fatalf("Error creating task: %v", err)
			}

			// Handle --tags flag
			tagsFlag, _ := cmd.Flags().GetString("tags")
			if tagsFlag != "" {
				tagNames := strings.Split(tagsFlag, ",")
				for _, name := range tagNames {
					name = strings.TrimSpace(name)
					if name == "" {
						continue
					}
					tag, err := dataService.TagsRepo.GetOrCreate(name)
					if err != nil {
						log.Printf("Error creating tag '%s': %v", name, err)
						continue
					}
					if err := dataService.TagsRepo.AttachToTask(tag.ID, task.ID); err != nil {
						log.Printf("Error attaching tag '%s' to task: %v", name, err)
					}
				}
			}
			fmt.Printf("Task created with ID: %d\n", task.ID)
		},
	}
	addCmd.Flags().String("tags", "", "Comma-separated list of tags (e.g., --tags=work,urgent)")

	// Tag management parent command
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	}

	// Tag create command
	tagCreateCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new tag",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			color, _ := cmd.Flags().GetString("color")
			desc, _ := cmd.Flags().GetString("desc")

			// If no color provided, use random from palette
			if color == "" {
				// Get count of existing tags to pick a color
				existingTags, _ := dataService.TagsRepo.GetAll()
				color = tasks.TagColorPalette[len(existingTags)%len(tasks.TagColorPalette)]
			}

			tag := &tasks.Tag{
				Name:        name,
				Color:       color,
				Description: desc,
			}

			if err := dataService.TagsRepo.Create(tag); err != nil {
				log.Fatalf("Error creating tag: %v", err)
			}
			fmt.Printf("Tag '%s' created with color %s\n", tag.Name, tag.Color)
		},
	}
	tagCreateCmd.Flags().String("color", "", "Tag color in hex format (e.g., #ff5733)")
	tagCreateCmd.Flags().String("desc", "", "Tag description")

	// Tag list command
	tagListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all tags",
		Run: func(cmd *cobra.Command, args []string) {
			tagsList, err := dataService.TagsRepo.GetAll()
			if err != nil {
				log.Fatalf("Error retrieving tags: %v", err)
			}
			if len(tagsList) == 0 {
				fmt.Println("No tags found")
				return
			}
			fmt.Println("Tags:")
			for _, tag := range tagsList {
				desc := ""
				if tag.Description != "" {
					desc = fmt.Sprintf(" - %s", tag.Description)
				}
				fmt.Printf("  %s (color: %s)%s\n", tag.Name, tag.Color, desc)
			}
		},
	}

	// Tag delete command
	tagDeleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a tag",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if err := dataService.TagsRepo.DeleteByName(name); err != nil {
				log.Fatalf("Error deleting tag: %v", err)
			}
			fmt.Printf("Tag '%s' deleted\n", name)
		},
	}

	tagCmd.AddCommand(tagCreateCmd, tagListCmd, tagDeleteCmd)

	// Task management parent command
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}

	// Task tag command (attach tag to task)
	taskTagCmd := &cobra.Command{
		Use:   "tag [task_id] [tag_name]",
		Short: "Attach a tag to a task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			taskID, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				log.Fatalf("Invalid task ID: %v", err)
			}
			tagName := args[1]

			// Get or create the tag
			tag, err := dataService.TagsRepo.GetOrCreate(tagName)
			if err != nil {
				log.Fatalf("Error getting/creating tag: %v", err)
			}

			// Verify task exists
			if _, err := dataService.TasksRepo.Get(uint(taskID)); err != nil {
				log.Fatalf("Task not found: %v", err)
			}

			if err := dataService.TagsRepo.AttachToTask(tag.ID, uint(taskID)); err != nil {
				log.Fatalf("Error attaching tag to task: %v", err)
			}
			fmt.Printf("Tag '%s' attached to task %d\n", tag.Name, taskID)
		},
	}

	// Task untag command (detach tag from task)
	taskUntagCmd := &cobra.Command{
		Use:   "untag [task_id] [tag_name]",
		Short: "Remove a tag from a task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			taskID, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				log.Fatalf("Invalid task ID: %v", err)
			}
			tagName := args[1]

			tag, err := dataService.TagsRepo.GetByName(tagName)
			if err != nil {
				log.Fatalf("Tag not found: %v", err)
			}

			if err := dataService.TagsRepo.DetachFromTask(tag.ID, uint(taskID)); err != nil {
				log.Fatalf("Error removing tag from task: %v", err)
			}
			fmt.Printf("Tag '%s' removed from task %d\n", tag.Name, taskID)
		},
	}

	taskCmd.AddCommand(taskTagCmd, taskUntagCmd)

	rootCmd.AddCommand(lsCmd, addCmd, tagCmd, taskCmd)
	return rootCmd
}
