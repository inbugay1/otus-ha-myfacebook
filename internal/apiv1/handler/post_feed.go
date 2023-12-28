package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"myfacebook/internal/apiv1"
	"myfacebook/internal/config"
	"myfacebook/internal/postfeedcache"
	"myfacebook/internal/repository"
)

type PostFeed struct {
	PostRepository repository.PostRepository
	UserRepository repository.UserRepository
	PostFeedCache  *postfeedcache.Cache
	EnvConfig      *config.EnvConfig
}

type postFeedRequest struct {
	Offset int
	Limit  int
}

func (h *PostFeed) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return apiv1.NewServerError(errUserIDTypeAssertionFailed)
	}

	postFeedReq, err := h.getPostFeedRequest(request)
	if err != nil {
		return err
	}

	lastRetrievedAtTimestamp, err := h.PostFeedCache.GetLastRetrievedAt(ctx, userID)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("post feed handler, failed to get last retrieved timestamp: %w", err))
	}

	if time.Since(time.UnixMilli(lastRetrievedAtTimestamp)) > time.Duration(h.EnvConfig.PopularFriendPostsRetrieveIntervalMinutes)*time.Minute {
		popularFriendsIDs, err := h.UserRepository.GetPopularFriendsIDsByUserID(ctx, userID, h.EnvConfig.PopularFriendUsersCount)
		if err != nil {
			return apiv1.NewServerError(fmt.Errorf("post feed handler, failed to get popular friends ids: %w", err))
		}

		if len(popularFriendsIDs) > 0 {
			popularFriendsPostsIDs, err := h.PostRepository.GetLastPostsIDsByAuthorIDs(ctx, popularFriendsIDs, lastRetrievedAtTimestamp, 1000)
			if err != nil {
				return apiv1.NewServerError(fmt.Errorf("post feed handler, failed to get last posts ids by author ids from post repo: %w", err))
			}

			for _, postID := range popularFriendsPostsIDs {
				err = h.PostFeedCache.AddPostID(ctx, userID, postID)
				if err != nil {
					return apiv1.NewServerError(fmt.Errorf("post feed handler, failed to add post id to post feed cache: %w", err))
				}
			}
		}

		err = h.PostFeedCache.SetLastRetrievedAt(ctx, userID, time.Now().UnixMilli())
		if err != nil {
			return apiv1.NewServerError(fmt.Errorf("post feed handler, failed to set last retrieved timestamp: %w", err))
		}
	}

	cachedPostsIDs, err := h.PostFeedCache.GetPostsIDs(ctx, userID)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("post feed handler, failed to fetch posts ids from cache: %w", err))
	}

	postFeedResponse := make([]getPostResponse, 0, postFeedReq.Limit)

	if len(cachedPostsIDs) > 0 {
		posts, err := h.PostRepository.GetPostsByIDs(ctx, cachedPostsIDs, postFeedReq.Offset, postFeedReq.Limit)
		if err != nil {
			return apiv1.NewServerError(fmt.Errorf("post feed handler, failed to get posts by ids from repo: %w", err))
		}

		for _, post := range posts {
			postFeedResponse = append(postFeedResponse, getPostResponse{
				ID:       post.ID,
				Text:     post.Text,
				AuthorID: post.AuthorID,
			})
		}
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	err = json.NewEncoder(responseWriter).Encode(&postFeedResponse)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("post feed handler, cannot encode response: %w", err))
	}

	return nil
}

func (h *PostFeed) getPostFeedRequest(request *http.Request) (postFeedRequest, error) {
	var postFeedReq postFeedRequest

	if request.URL.Query().Get("offset") == "" {
		postFeedReq.Offset = 0
	} else {
		offset, err := strconv.Atoi(request.URL.Query().Get("offset"))
		if err != nil {
			return postFeedReq, apiv1.NewInvalidRequestErrorInvalidParameter("offset",
				fmt.Errorf("post feed handler, failed to convert offset %q to int: %w", request.URL.Query().Get("offset"), err))
		}

		postFeedReq.Offset = offset
	}

	if request.URL.Query().Get("limit") == "" {
		postFeedReq.Limit = 10
	} else {
		limit, err := strconv.Atoi(request.URL.Query().Get("limit"))
		if err != nil {
			return postFeedReq, apiv1.NewInvalidRequestErrorInvalidParameter("limit",
				fmt.Errorf("post feed handler, failed to convert limit %q to int: %w", request.URL.Query().Get("limit"), err))
		}

		postFeedReq.Limit = limit
	}

	return postFeedReq, nil
}
