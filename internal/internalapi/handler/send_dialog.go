package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"myfacebook/internal/internalapi"
	"myfacebook/internal/repository"
)

type SendDialog struct {
	DialogRepository repository.DialogRepository
	UserRepository   repository.UserRepository
}

type sendDialogRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

func (h *SendDialog) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	var sendDialogReq sendDialogRequest
	if err := json.NewDecoder(request.Body).Decode(&sendDialogReq); err != nil {
		return internalapi.NewServerError(fmt.Errorf("send dialog handler, cannot decode request body: %w", err))
	}

	defer request.Body.Close()

	err := h.validateSendDialogRequest(ctx, sendDialogReq)
	if err != nil {
		return err
	}

	dialogMessage := repository.DialogMessage{
		From: sendDialogReq.From,
		To:   sendDialogReq.To,
		Text: sendDialogReq.Text,
	}

	err = h.DialogRepository.Add(ctx, dialogMessage)
	if err != nil {
		return internalapi.NewServerError(fmt.Errorf("send dialog handler, failed to add dialog message to repository: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	return nil
}

func (h *SendDialog) validateSendDialogRequest(ctx context.Context, sendDialogReq sendDialogRequest) error {
	if sendDialogReq.Text == "" {
		return internalapi.NewInvalidRequestErrorMissingRequiredParameter("text")
	}

	if sendDialogReq.From == "" {
		return internalapi.NewInvalidRequestErrorMissingRequiredParameter("from")
	}

	uuidv4Regexp := regexp.MustCompile(`(?i)^[a-f\d]{8}-[a-f\d]{4}-4[a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$`)
	if !uuidv4Regexp.MatchString(sendDialogReq.From) {
		return internalapi.NewInvalidRequestErrorInvalidParameter("from", nil)
	}

	_, err := h.UserRepository.GetUserByID(ctx, sendDialogReq.From)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return internalapi.NewInvalidRequestErrorInvalidParameter("from", err)
		}

		return internalapi.NewServerError(err)
	}

	if sendDialogReq.To == "" {
		return internalapi.NewInvalidRequestErrorMissingRequiredParameter("to")
	}

	if !uuidv4Regexp.MatchString(sendDialogReq.To) {
		return internalapi.NewInvalidRequestErrorInvalidParameter("to", nil)
	}

	_, err = h.UserRepository.GetUserByID(ctx, sendDialogReq.To)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return internalapi.NewInvalidRequestErrorInvalidParameter("to", err)
		}

		return internalapi.NewServerError(err)
	}

	return nil
}
