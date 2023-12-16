package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"myfacebook/internal/apiv1"
	"myfacebook/internal/repository"
)

type Register struct {
	UserRepository repository.UserRepository
}

type registerRequest struct {
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	Birthdate  string `json:"birthdate"`
	Biography  string `json:"biography"`
	City       string `json:"city"`
	Password   string `json:"password"`
}

type registerResponse struct {
	UserID string `json:"user_id"`
}

func (h *Register) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	var registerReq registerRequest
	if err := json.NewDecoder(request.Body).Decode(&registerReq); err != nil {
		return apiv1.NewServerError(fmt.Errorf("register handler, cannot decode request body: %w", err))
	}

	defer request.Body.Close()

	err := h.validateRegisterRequest(registerReq)
	if err != nil {
		return err
	}

	userUUIDv4, err := uuid.NewV4()
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("register handler, failed to generate user uuid: %w", err))
	}

	user := repository.User{
		ID:        userUUIDv4.String(),
		FirstName: registerReq.FirstName,
		LastName:  registerReq.SecondName,
		BirthDate: registerReq.Birthdate,
		Biography: registerReq.Biography,
		City:      registerReq.City,
		Password:  hashPassword(registerReq.Password),
	}

	ctx := request.Context()

	if err := h.UserRepository.Add(ctx, user); err != nil {
		return apiv1.NewServerError(fmt.Errorf("register handler, failed to add user to repository: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	err = json.NewEncoder(responseWriter).Encode(registerResponse{
		UserID: user.ID,
	})
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("register handler, cannot encode response: %w", err))
	}

	return nil
}

func (h *Register) validateRegisterRequest(registerReq registerRequest) error {
	if registerReq.FirstName == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("first_name")
	}

	if registerReq.SecondName == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("second_name")
	}

	if registerReq.Birthdate == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("birthdate")
	}

	_, err := time.Parse("2006-01-02", registerReq.Birthdate)
	if err != nil {
		return apiv1.NewInvalidRequestErrorInvalidParameter("birthdate", err)
	}

	if registerReq.City == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("city")
	}

	if registerReq.Password == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("password")
	}

	const passwordLen = 6

	if len(registerReq.Password) < passwordLen {
		return apiv1.NewInvalidRequestError(fmt.Sprintf("min passsword len is %d", passwordLen), nil)
	}

	return nil
}

func hashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))

	return hex.EncodeToString(hash.Sum(nil))
}
