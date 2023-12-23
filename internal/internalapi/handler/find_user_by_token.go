package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/internalapi"
	"myfacebook/internal/repository"
)

type FindUserByToken struct {
	UserRepository repository.UserRepository
}

type findUserByTokenResponse struct {
	ID string `json:"id"`
}

func (h *FindUserByToken) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	token := httprouter.RouteParam(ctx, "token")

	user, err := h.UserRepository.GetUserByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return internalapi.NewEntityNotFoundError(err)
		}

		return internalapi.NewServerError(fmt.Errorf("find user by token handler, failed to get user by token from repository: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	err = json.NewEncoder(responseWriter).Encode(findUserByTokenResponse{
		ID: user.ID,
	})
	if err != nil {
		return internalapi.NewServerError(fmt.Errorf("find user by token handler, cannot encode response: %w", err))
	}

	return nil
}
