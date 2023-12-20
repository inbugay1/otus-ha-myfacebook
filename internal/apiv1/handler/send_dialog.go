package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
)

type SendDialog struct {
	DialogRepository repository.DialogRepository
}

type sendDialogRequest struct {
	Text string `json:"text"`
}

func (h *SendDialog) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	var sendDialogReq sendDialogRequest
	if err := json.NewDecoder(request.Body).Decode(&sendDialogReq); err != nil {
		return apiv1.NewServerError(fmt.Errorf("send dialogMessage handler, cannot decode request body: %w", err))
	}

	defer request.Body.Close()

	if sendDialogReq.Text == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("text")
	}

	ctx := request.Context()

	senderID := ctx.Value("user_id").(string)
	receiverID := httprouter.RouteParam(ctx, "user_id")

	dialogMessage := repository.DialogMessage{
		From: senderID,
		To:   receiverID,
		Text: sendDialogReq.Text,
	}

	err := h.DialogRepository.Add(ctx, dialogMessage)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("send dialog handler, failed to add dialog message to repository: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	return nil
}
