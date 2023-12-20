package handler

import (
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
)

type SetFriend struct {
	UserRepository repository.UserRepository
}

func (h *SetFriend) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	userID := ctx.Value("user_id").(string)

	friendID := httprouter.RouteParam(ctx, "id")

	err := h.UserRepository.SetUserFriend(ctx, userID, friendID)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("set friend handler, failed to set user friend: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	return nil
}
