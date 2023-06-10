package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type NotificationClient struct {
	HttpClient *http.Client
	WebhookUrl string
}

func NewNotificationClient(webhookUrl string) *NotificationClient {
	return &NotificationClient{HttpClient: http.DefaultClient, WebhookUrl: webhookUrl}
}

type NotificationRequestBody struct {
	Text string `json:"text"`
}

func (c *NotificationClient) NotifyInsertTaskError(task OismTask, err error) error {
	return c.notify(fmt.Sprintf("タスク: %vをリスト: %v に追加できませんでした。Error: %v extra: %v", task.Title, task.ListName, err.Error(), strings.Join([]string{task.Notes, task.Duo}, ",")))
}

func (c *NotificationClient) notify(content string) error {
	j, err := json.Marshal(NotificationRequestBody{Text: content})
	if err != nil {
		fmt.Println(err)
	}
	req, err := http.NewRequest("POST", c.WebhookUrl, bytes.NewBuffer(j))
	if err != nil {
		fmt.Println(err)
	}
	_, err = c.HttpClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
