package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/usecase/login"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/usecase/refresh"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/usecase/register"
	"github.com/google/uuid"
)

type AuthHandler struct {
	registerUC *register.UseCase
	loginUC    *login.UseCase
	refreshUC  *refresh.UseCase
}

func NewAuthHandler(
	registerUC *register.UseCase,
	loginUC *login.UseCase,
	refreshUC *refresh.UseCase,
) *AuthHandler {
	return &AuthHandler{
		registerUC: registerUC,
		loginUC:    loginUC,
		refreshUC:  refreshUC,
	}
}

// DTOs de Request
type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func renderError(w http.ResponseWriter, status int, message string) {
	renderJSON(w, status, errorResponse{Error: message})
}

// Register trata POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	input := register.InputDTO{
		ID:       uuid.NewString(),
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     "customer",
	}

	output, err := h.registerUC.Execute(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, register.ErrUserAlreadyExists):
			renderError(w, http.StatusConflict, err.Error())
		case errors.Is(err, register.ErrPasswordTooShort),
			errors.Is(err, entity.ErrEmptyName),
			errors.Is(err, entity.ErrInvalidEmail):
			renderError(w, http.StatusBadRequest, err.Error())
		default:
			renderError(w, http.StatusInternalServerError, "erro interno ao cadastrar usuário")
		}
		return
	}

	renderJSON(w, http.StatusCreated, output)
}

// Login trata POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	input := login.InputDTO{
		Email:    req.Email,
		Password: req.Password,
	}

	output, err := h.loginUC.Execute(r.Context(), input)
	if err != nil {
		if errors.Is(err, login.ErrInvalidCredentials) {
			renderError(w, http.StatusUnauthorized, "email ou senha incorretos")
			return
		}
		renderError(w, http.StatusInternalServerError, "erro interno ao autenticar")
		return
	}

	renderJSON(w, http.StatusOK, output)
}

// Refresh trata POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	input := refresh.InputDTO{
		RefreshToken: req.RefreshToken,
	}

	output, err := h.refreshUC.Execute(r.Context(), input)
	if err != nil {
		if errors.Is(err, refresh.ErrInvalidRefreshToken) {
			renderError(w, http.StatusUnauthorized, "refresh token inválido ou expirado")
			return
		}
		renderError(w, http.StatusInternalServerError, "erro interno ao renovar token")
		return
	}

	renderJSON(w, http.StatusOK, output)
}

// HealthCheck trata GET /health
func (h *AuthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	renderJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "auth-service",
	})
}
