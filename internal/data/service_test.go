package data

import (
	"testing"

	"github.com/mustafmst/taskerr/internal/config"
)

func TestCreateAndUpdateTaskCentralizesTagManagement(t *testing.T) {
	service, err := NewService(&config.Config{
		DBProvider:   "sqlite",
		DBConnection: t.TempDir() + "/taskerr.db",
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	workTag, err := service.TagsRepo.GetOrCreate("work")
	if err != nil {
		t.Fatalf("GetOrCreate(work) error = %v", err)
	}

	task, err := service.CreateTask("ship release", "document the rollout", []uint{workTag.ID}, []string{"urgent"})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if task.Details != "document the rollout" {
		t.Fatalf("Details = %q, want details to persist", task.Details)
	}

	taskTags, err := service.TagsRepo.GetTaskTags(task.ID)
	if err != nil {
		t.Fatalf("GetTaskTags(create) error = %v", err)
	}
	if len(taskTags) != 2 {
		t.Fatalf("len(taskTags) after create = %d, want 2", len(taskTags))
	}

	personalTag, err := service.TagsRepo.GetOrCreate("personal")
	if err != nil {
		t.Fatalf("GetOrCreate(personal) error = %v", err)
	}

	if err := service.UpdateTask(task.ID, "ship release notes", "write changelog and email", []uint{personalTag.ID}, nil); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	updatedTask, err := service.TasksRepo.Get(task.ID)
	if err != nil {
		t.Fatalf("TasksRepo.Get() error = %v", err)
	}
	if updatedTask.Description != "ship release notes" {
		t.Fatalf("Description = %q, want updated value", updatedTask.Description)
	}
	if updatedTask.Details != "write changelog and email" {
		t.Fatalf("Details = %q, want updated value", updatedTask.Details)
	}

	taskTags, err = service.TagsRepo.GetTaskTags(task.ID)
	if err != nil {
		t.Fatalf("GetTaskTags(update) error = %v", err)
	}
	if len(taskTags) != 1 {
		t.Fatalf("len(taskTags) after update = %d, want 1", len(taskTags))
	}
	if taskTags[0].ID != personalTag.ID {
		t.Fatalf("remaining tag ID = %d, want %d", taskTags[0].ID, personalTag.ID)
	}
}

func TestAttachAndDetachTagToTask(t *testing.T) {
	service, err := NewService(&config.Config{
		DBProvider:   "sqlite",
		DBConnection: t.TempDir() + "/taskerr.db",
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	task, err := service.CreateTask("review code", "", nil, nil)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	tag, err := service.AttachTagToTask(task.ID, "backend")
	if err != nil {
		t.Fatalf("AttachTagToTask() error = %v", err)
	}

	taskTags, err := service.TagsRepo.GetTaskTags(task.ID)
	if err != nil {
		t.Fatalf("GetTaskTags(attach) error = %v", err)
	}
	if len(taskTags) != 1 || taskTags[0].ID != tag.ID {
		t.Fatalf("unexpected tags after attach: %+v", taskTags)
	}

	if err := service.DetachTagFromTask(task.ID, "backend"); err != nil {
		t.Fatalf("DetachTagFromTask() error = %v", err)
	}

	taskTags, err = service.TagsRepo.GetTaskTags(task.ID)
	if err != nil {
		t.Fatalf("GetTaskTags(detach) error = %v", err)
	}
	if len(taskTags) != 0 {
		t.Fatalf("len(taskTags) after detach = %d, want 0", len(taskTags))
	}
}
