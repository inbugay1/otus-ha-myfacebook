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

type GetUser struct {
	UserRepository repository.UserRepository
}

type getUserResponse struct {
	ID         string `json:"id"`
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	Birthdate  string `json:"birthdate"`
	Biography  string `json:"biography"`
	City       string `json:"city"`
}

func (h *GetUser) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	userID := httprouter.RouteParam(ctx, "id")

	user, err := h.UserRepository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apiv1.NewEntityNotFoundError(err)
		}

		return apiv1.NewServerError(fmt.Errorf("get user handler, failed to get user by id from repository: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	err = json.NewEncoder(responseWriter).Encode(getUserResponse{
		ID:         user.ID,
		FirstName:  user.FirstName,
		SecondName: user.LastName,
		Birthdate:  user.BirthDate,
		Biography:  user.Biography,
		City:       user.City,
	})
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("get user handler, cannot encode response: %w", err))
	}

	return nil
}
