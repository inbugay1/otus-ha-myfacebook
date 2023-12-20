package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
)

type SearchUser struct {
	UserRepository repository.UserRepository
}

type searchUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type userResponse struct {
	ID         string `json:"id"`
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	Birthdate  string `json:"birthdate"`
	Biography  string `json:"biography"`
	City       string `json:"city"`
}

func (h *SearchUser) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	searchUserReq := h.getSearchRequest(request)

	if err := h.validateRequest(searchUserReq); err != nil {
		return err
	}

	ctx := request.Context()

	users, err := h.UserRepository.GetUsersByFirstnameAndLastname(ctx, searchUserReq.FirstName, searchUserReq.LastName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apiv1.NewEntityNotFoundError(err)
		}

		return apiv1.NewServerError(fmt.Errorf("search user handler, failed to fetch user from repository: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	searchUserResponse := make([]userResponse, 0, len(users))

	for _, user := range users {
		searchUserResponse = append(searchUserResponse, userResponse{
			ID:         user.ID,
			FirstName:  user.FirstName,
			SecondName: user.LastName,
			Birthdate:  user.BirthDate,
			Biography:  user.Biography,
			City:       user.City,
		})
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	err = json.NewEncoder(responseWriter).Encode(&searchUserResponse)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("search user handler, cannot encode response: %w", err))
	}

	return nil
}

func (h *SearchUser) getSearchRequest(request *http.Request) searchUserRequest {
	searchUserReq := searchUserRequest{
		FirstName: request.URL.Query().Get("first_name"),
		LastName:  request.URL.Query().Get("last_name"),
	}

	return searchUserReq
}

func (h *SearchUser) validateRequest(searchUserReq searchUserRequest) error {
	if searchUserReq.FirstName == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("first_name")
	}

	if searchUserReq.LastName == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("last_name")
	}

	return nil
}
