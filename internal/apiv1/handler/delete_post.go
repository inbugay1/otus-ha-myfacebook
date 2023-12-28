package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
	"myfacebook/internal/rmq"
)

type DeletePost struct {
	PostRepository repository.PostRepository
	RMQ            *rmq.RMQ
}

func (h *DeletePost) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	authorID, ok := ctx.Value("user_id").(string)
	if !ok {
		return apiv1.NewServerError(errUserIDTypeAssertionFailed)
	}

	postID := httprouter.RouteParam(ctx, "id")

	err := h.PostRepository.Delete(ctx, postID, authorID)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("delete post handler, failed to delete post: %w", err))
	}

	rmqMessage, err := json.Marshal(postMessage{
		Operation: "remove",
		PostID:    postID,
		AuthorID:  authorID,
	})
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("delete post handler, failed to make rmq message: %w", err))
	}

	err = h.RMQ.Publish(ctx, "postfeed", rmqMessage)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("delete post handler, failed to publish rmq message: %w", err))
	}

	responseWriter.WriteHeader(http.StatusOK)

	return nil
}
