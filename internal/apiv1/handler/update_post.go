package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
)

type UpdatePost struct {
	PostRepository repository.PostRepository
}

type updatePostRequest struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func (h *UpdatePost) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	var updatePostReq updatePostRequest
	if err := json.NewDecoder(request.Body).Decode(&updatePostReq); err != nil {
		return apiv1.NewServerError(fmt.Errorf("update post handler, cannot decode request body: %w", err))
	}

	defer request.Body.Close()

	if updatePostReq.ID == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("id")
	}

	if updatePostReq.Text == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("text")
	}

	post := repository.Post{
		ID:   updatePostReq.ID,
		Text: updatePostReq.Text,
	}

	err := h.PostRepository.Update(ctx, post)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apiv1.NewEntityNotFoundError(err)
		}

		return apiv1.NewServerError(fmt.Errorf("update post handler, failed to get post from repo: %w", err))
	}

	responseWriter.WriteHeader(http.StatusOK)

	return nil
}
