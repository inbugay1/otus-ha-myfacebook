package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
	"myfacebook/internal/repository/rest"
	sqlxrepo "myfacebook/internal/repository/sqlx"
)

type SendDialog struct {
	SqlxDialogRepository *sqlxrepo.DialogRepository
	RestDialogRepository *rest.DialogRepository
}

type sendDialogRequest struct {
	Text string `json:"text"`
}

func (h *SendDialog) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	var sendDialogReq sendDialogRequest
	if err := json.NewDecoder(request.Body).Decode(&sendDialogReq); err != nil {
		return apiv1.NewServerError(fmt.Errorf("send dialogMessage handler cannot decode request body: %w", err))
	}

	defer request.Body.Close()

	if sendDialogReq.Text == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("text")
	}

	ctx := request.Context()

	senderID := ctx.Value("user_id").(string)
	receiverID := httprouter.RouteParam(ctx, "user_id")

	dialogMsg := repository.DialogMessage{
		From: senderID,
		To:   receiverID,
		Text: sendDialogReq.Text,
	}

	err := h.RestDialogRepository.Add(ctx, dialogMsg)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("send dialog handler failed to add dialog message to rest repository: %w", err))
	}

	err = h.SqlxDialogRepository.Add(ctx, dialogMsg)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("send dialog handler failed to add dialog message to sqlx repository: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	return nil
}
