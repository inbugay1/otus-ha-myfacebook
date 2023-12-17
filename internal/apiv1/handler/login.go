package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gofrs/uuid"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
)

type Login struct {
	UserRepository repository.UserRepository
}

type loginRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *Login) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	var loginReq loginRequest
	if err := json.NewDecoder(request.Body).Decode(&loginReq); err != nil {
		return apiv1.NewServerError(fmt.Errorf("login handler, cannot decode request body: %w", err))
	}

	defer request.Body.Close()

	if err := h.validateLoginRequest(loginReq); err != nil {
		return err
	}

	ctx := request.Context()

	user, err := h.UserRepository.GetUserByID(ctx, loginReq.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apiv1.NewEntityNotFoundError(err)
		}

		return apiv1.NewServerError(fmt.Errorf("login handler, failed to get user by id: %w", err))
	}

	if user == nil || hashPassword(loginReq.Password) != user.Password {
		return apiv1.NewInvalidCredentialsError()
	}

	tokenUUIDv4, err := uuid.NewV4()
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("login handler, failed to generate token uuid: %w", err))
	}

	token := tokenUUIDv4.String()

	err = h.UserRepository.UpdateUserToken(ctx, user.ID, token)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("login handler, failed to update user token: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(responseWriter).Encode(loginResponse{Token: token}); err != nil {
		return apiv1.NewServerError(fmt.Errorf("login handler, cannot encode response: %w", err))
	}

	return nil
}

func (h *Login) validateLoginRequest(loginReq loginRequest) error {
	if loginReq.ID == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("id")
	}

	uuidv4Regexp := regexp.MustCompile(`(?i)^[a-f\d]{8}-[a-f\d]{4}-4[a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$`)
	if !uuidv4Regexp.MatchString(loginReq.ID) {
		return apiv1.NewInvalidRequestErrorInvalidParameter("id", nil)
	}

	if loginReq.Password == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("password")
	}

	return nil
}
