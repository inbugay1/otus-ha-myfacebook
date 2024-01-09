package handler

import (
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/postfeedcache"
	"myfacebook/internal/repository"
)

type DeleteFriend struct {
	UserRepository repository.UserRepository
	PostRepository repository.PostRepository
	PostFeedCache  *postfeedcache.Cache
}

func (h *DeleteFriend) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	userID := ctx.Value("user_id").(string)

	friendID := httprouter.RouteParam(ctx, "id")

	err := h.UserRepository.DeleteFriend(ctx, userID, friendID)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("delete friend handler, failed to delete friend: %w", err))
	}

	postsIDs, err := h.PostFeedCache.GetPostsIDs(ctx, userID)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("delete friend handler, failed to get posts ids from post feed: %w", err))
	}

	if len(postsIDs) > 0 {
		posts, err := h.PostRepository.GetPostsByIDs(ctx, postsIDs, 0, 1000)
		if err != nil {
			return apiv1.NewServerError(fmt.Errorf("delete friend handler, failed to get posts by ids from repo: %w", err))
		}

		for _, post := range posts {
			if post.AuthorID == friendID {
				err := h.PostFeedCache.RemovePostID(ctx, userID, post.ID)
				if err != nil {
					return apiv1.NewServerError(fmt.Errorf("delete friend handler, failed to remove post from post feed: %w", err))
				}
			}
		}
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	return nil
}
