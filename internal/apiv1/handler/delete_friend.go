package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
)

type DeleteFriend struct {
	UserRepository repository.UserRepository
}

func (h *DeleteFriend) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	userID := ctx.Value("user_id").(string)

	friendID := httprouter.RouteParam(ctx, "id")

	user, err := h.UserRepository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apiv1.NewEntityNotFoundError(err)
		}

		return apiv1.NewServerError(fmt.Errorf("delete friend handler, failed to get user by id: %w", err))
	}

	if user.FriendID != friendID {
		return apiv1.NewInvalidRequestError("invalid friend id", nil)
	}

	err = h.UserRepository.SetUserFriend(ctx, userID, "")
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("delete friend handler, failed to set user friend: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	return nil
}
