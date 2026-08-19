package services

import (
	"context"
	"relay/internal/domain"
	"testing"
)

func TestMessageService_Send_EmptyMessage(t *testing.T) {
	service := &MessageService{}

	_, err := service.Send(
		context.Background(),
		1,
		1,
		"  ",
	)

	if err != domain.ErrEmptyMessage {
		t.Fatalf(
			"expected ErrEmptyMessage, got %v",
			err,
		)
	}
}
