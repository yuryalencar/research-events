package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/yuryalencar/research-events/internal/middleware"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/service"
)

// --- Types ---

// UserHandler handles user self-service endpoints.
// Dependencies are injected via constructor — never use globals.
type UserHandler struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
}

// --- Constructor ---

func NewUserHandler(userRepo repository.UserRepository, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// --- Public methods ---

// UpdatePassword handles PATCH /api/v1/users/me/password.
// Requires a valid JWT (RequireAuth middleware must run before this handler).
// Validation order matches the spec:
//  1. Auth user present in context (JWT guard)
//  2. JSON body parseable
//  3. Fields non-empty + passwords match
//  4. New password complexity
//  5. Current password verified against DB
//  6. Hash new password and persist
func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "TOKEN_MISSING", "authentication token is missing")
		return
	}

	var input struct {
		CurrentPassword         string `json:"current_password"`
		NewPassword             string `json:"new_password"`
		NewPasswordConfirmation string `json:"new_password_confirmation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	if err := service.ValidateUpdatePasswordInput(input.CurrentPassword, input.NewPassword, input.NewPasswordConfirmation); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := service.ValidatePasswordComplexity(input.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), authUser.ID)
	if err != nil {
		h.logger.Error("failed to find user for password update", "user_id", authUser.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	if err := service.ValidateCredentials(user.PasswordHash, input.CurrentPassword); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CURRENT_PASSWORD", "current password is incorrect")
		return
	}

	newHash, err := service.HashPassword(input.NewPassword)
	if err != nil {
		h.logger.Error("failed to hash new password", "user_id", authUser.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	if err := h.userRepo.UpdatePassword(r.Context(), authUser.ID, newHash); err != nil {
		h.logger.Error("failed to update password", "user_id", authUser.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeSuccess(w, http.StatusOK, "PASSWORD_UPDATED", map[string]string{
		"message": "Password updated successfully.",
	})
}
