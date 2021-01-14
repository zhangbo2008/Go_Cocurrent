package handlers

import (
	"context"
	"fmt"

	v1 "awesomeProject8/myproject/pkg/apis/v1"
)

// ListMessages returns all messages.
func ListMessages(ctx context.Context, count int) ([]v1.Message, error) {
	messages := make([]v1.Message, count)
	for i := 0; i < count; i++ {
		messages[i].ID = i
		messages[i].Title = fmt.Sprintf("Example %d", i)
		messages[i].Content = fmt.Sprintf("Content of example %d", i)
	}
	return messages, nil
}

// GetMessage return a message by id.
func GetMessage(ctx context.Context, id int) (*v1.Message, error) {
	return &v1.Message{
		ID:      id,
		Title:   "This is an example",
		Content: "Example content",
	}, nil
}
