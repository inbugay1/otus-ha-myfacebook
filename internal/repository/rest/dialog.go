package rest

import (
	"context"
	"fmt"

	"myfacebook/internal/myfacebookdialogapiclient"
	"myfacebook/internal/repository"
)

type DialogRepository struct {
	apiClient *myfacebookdialogapiclient.Client
}

func NewDialogRepository(apiClient *myfacebookdialogapiclient.Client) *DialogRepository {
	return &DialogRepository{
		apiClient: apiClient,
	}
}

func (r *DialogRepository) Add(ctx context.Context, dialogMessage repository.DialogMessage) error {
	err := r.apiClient.SendDialogMessage(ctx, dialogMessage.From, dialogMessage.To, dialogMessage.Text)
	if err != nil {
		return fmt.Errorf("rest dialogrepository failed to send dialog message: %w", err)
	}

	return nil
}

func (r *DialogRepository) GetDialogMessagesBySenderIDAndReceiverID(ctx context.Context, senderID, receiverID string) ([]repository.DialogMessage, error) {
	dialogMessages, err := r.apiClient.GetDialogMessages(ctx, senderID, receiverID)
	if err != nil {
		return nil, fmt.Errorf("rest dialogrepository failed to get dialog messages: %w", err)
	}

	res := make([]repository.DialogMessage, 0, len(dialogMessages))

	for _, dialogMessage := range dialogMessages {
		res = append(res, repository.DialogMessage{
			ID:   dialogMessage.ID,
			From: dialogMessage.From,
			To:   dialogMessage.To,
			Text: dialogMessage.Text,
		})
	}

	return res, nil
}
