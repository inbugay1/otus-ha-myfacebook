package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
)

type ListDialog struct {
	DialogRepository repository.DialogRepository
}

type dialogMessage struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

func (h *ListDialog) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	senderID := ctx.Value("user_id").(string)
	receiverID := httprouter.RouteParam(ctx, "user_id")

	dialogMessages, err := h.DialogRepository.GetDialogMessagesBySenderIDAndReceiverID(ctx, senderID, receiverID)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("list dialog handler, failed to fetch dialoag messages from repository: %w", err))
	}

	listDialogResponse := make([]dialogMessage, 0, len(dialogMessages))

	for _, dialogMsg := range dialogMessages {
		listDialogResponse = append(listDialogResponse, dialogMessage{
			From: dialogMsg.From,
			To:   dialogMsg.To,
			Text: dialogMsg.Text,
		})
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	err = json.NewEncoder(responseWriter).Encode(&listDialogResponse)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("list dialog handler, cannot encode response: %w", err))
	}

	return nil
}
