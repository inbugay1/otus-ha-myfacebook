package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
)

type GetPost struct {
	PostRepository repository.PostRepository
}

type getPostResponse struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	AuthorID string `json:"author_id"`
}

func (h *GetPost) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	postID := httprouter.RouteParam(ctx, "id")

	post, err := h.PostRepository.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apiv1.NewEntityNotFoundError(err)
		}

		return apiv1.NewServerError(fmt.Errorf("get post handler, failed to get post from repo: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	err = json.NewEncoder(responseWriter).Encode(getPostResponse{
		ID:       post.ID,
		Text:     post.Text,
		AuthorID: post.AuthorID,
	})
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("get post handler, cannot encode response: %w", err))
	}

	return nil
}
