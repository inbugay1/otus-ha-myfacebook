package repository

import "context"

type DialogMessage struct {
	ID   string `db:"id"`
	From string `db:"sender_id"`
	To   string `db:"receiver_id"`
	Text string `db:"text"`
}

type DialogRepository interface {
	Add(ctx context.Context, dialog DialogMessage) error
	GetDialogMessagesBySenderIDAndReceiverID(ctx context.Context, senderID, receiverID string) ([]DialogMessage, error)
}
