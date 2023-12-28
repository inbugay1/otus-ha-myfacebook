package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
	"myfacebook/internal/rmq"
)

var errUserIDTypeAssertionFailed = errors.New("failed to assert user_id to string")

type CreatePost struct {
	PostRepository repository.PostRepository
	RMQ            *rmq.RMQ
}

type postRequest struct {
	Text string `json:"text"`
}

type postResponse struct {
	ID string `json:"id"`
}

type postMessage struct {
	Operation string `json:"operation"`
	PostID    string `json:"post_id"`
	AuthorID  string `json:"author_id"`
}

func (h *CreatePost) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	var postReq postRequest
	if err := json.NewDecoder(request.Body).Decode(&postReq); err != nil {
		return apiv1.NewServerError(fmt.Errorf("create post handler, cannot decode request body: %w", err))
	}

	defer request.Body.Close()

	if postReq.Text == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("text")
	}

	ctx := request.Context()

	authorID, ok := ctx.Value("user_id").(string)
	if !ok {
		return apiv1.NewServerError(errUserIDTypeAssertionFailed)
	}

	postUUIDv4, err := uuid.NewV4()
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("create post handler, failed to generate user uuid: %w", err))
	}

	post := repository.Post{
		ID:       postUUIDv4.String(),
		Text:     postReq.Text,
		AuthorID: authorID,
	}

	err = h.PostRepository.Add(ctx, post)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("create post handler, failed to add post: %w", err))
	}

	rmqMessage, err := json.Marshal(postMessage{
		Operation: "add",
		PostID:    post.ID,
		AuthorID:  authorID,
	})
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("create post handler, failed to make rmq message: %w", err))
	}

	err = h.RMQ.Publish(ctx, "postfeed", rmqMessage)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("create post handler, failed to publish rmq message: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	err = json.NewEncoder(responseWriter).Encode(postResponse{
		ID: post.ID,
	})
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("create post handler, cannot encode response: %w", err))
	}

	return nil
}
