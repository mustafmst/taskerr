package cli

import (
	"fmt"
	"log"

	"github.com/mustafmst/taskerr/internal/data"
	"github.com/mustafmst/taskerr/internal/data/tasks"
	"github.com/spf13/cobra"
)

func newTagCommand(dataService *data.Service) *cobra.Command {
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	}

	tagCmd.AddCommand(
		newTagCreateCommand(dataService),
		newTagListCommand(dataService),
		newTagDeleteCommand(dataService),
	)

	return tagCmd
}

func newTagCreateCommand(dataService *data.Service) *cobra.Command {
	tagCreateCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new tag",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			color, _ := cmd.Flags().GetString("color")
			desc, _ := cmd.Flags().GetString("desc")

			if color == "" {
				existingTags, err := dataService.TagsRepo.GetAll()
				if err != nil {
					log.Fatalf("Error retrieving tags: %v", err)
				}
				color = tasks.TagColorPalette[len(existingTags)%len(tasks.TagColorPalette)]
			}

			tag := &tasks.Tag{
				Name:        args[0],
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
	return tagCreateCmd
}

func newTagListCommand(dataService *data.Service) *cobra.Command {
	return &cobra.Command{
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
}

func newTagDeleteCommand(dataService *data.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a tag",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := dataService.TagsRepo.DeleteByName(args[0]); err != nil {
				log.Fatalf("Error deleting tag: %v", err)
			}

			fmt.Printf("Tag '%s' deleted\n", args[0])
		},
	}
}
