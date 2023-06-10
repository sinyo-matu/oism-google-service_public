package main

import (
	"context"
	"sync"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

type TasksService struct {
	inner *tasks.Service
}

var TaskListNameIdStore sync.Map

func NewTasksService(token *oauth2.Token) (*TasksService, error) {
	client := AuthConfig.Client(context.Background(), token)
	s, err := tasks.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return &TasksService{inner: s}, nil
}

func (s *TasksService) InsertTask(taskListName, title, notes, duo string) error {
	targetId_i, found := TaskListNameIdStore.Load(taskListName)
	if found {
		targetId := targetId_i.(string)
		newTask := tasks.Task{Title: title, Notes: notes, Due: duo}
		_, err := s.inner.Tasks.Insert(targetId, &newTask).Do()
		if err != nil {
			return err
		}
		return nil
	}
	taskList, err := s.inner.Tasklists.List().Do()
	if err != nil {
		return err
	}
	targetId := ""
	for _, l := range taskList.Items {
		if l.Title == taskListName {
			targetId = l.Id
			TaskListNameIdStore.Store(taskListName, l.Id)
		}
	}
	if targetId == "" {
		newList := tasks.TaskList{Title: taskListName}
		list, err := s.inner.Tasklists.Insert(&newList).Do()
		targetId = list.Id
		TaskListNameIdStore.Store(taskListName, list.Id)
		if err != nil {
			return err
		}
	}
	newTask := tasks.Task{Title: title, Notes: notes, Due: duo}
	_, err = s.inner.Tasks.Insert(targetId, &newTask).Do()
	if err != nil {
		return err
	}
	return nil
}
