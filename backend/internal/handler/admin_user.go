package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/yuryalencar/research-events/internal/middleware"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/service"
)

// --- Types ---

// AdminUserHandler handles admin-scoped user management endpoints.
// Must be protected by RequireAuth + RequireRole("admin") in the route registration.
type AdminUserHandler struct {
	userRepo  repository.UserRepository
	auditRepo repository.AuditRepository
	logger    *slog.Logger
}

// --- Constructor ---

func NewAdminUserHandler(
	userRepo repository.UserRepository,
	auditRepo repository.AuditRepository,
	logger *slog.Logger,
) *AdminUserHandler {
	return &AdminUserHandler{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// --- Public methods ---

// Unlock handles PATCH /api/v1/admin/users/{id}/unlock.
// Requires admin role — enforced at route level via RequireAuth + RequireRole("admin").
func (h *AdminUserHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	// 1. Parse target user ID from the URL path parameter.
	rawID := r.PathValue("id")
	targetID64, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", ":id must be a valid integer")
		return
	}
	targetID := uint(targetID64)

	// 2. Get the authenticated admin from context (set by RequireAuth middleware).
	admin, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		// Should never happen when RequireAuth is in the chain — guard for safety.
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	// 3. Admins cannot unlock their own account.
	if targetID == admin.ID {
		writeError(w, http.StatusUnprocessableEntity, "CANNOT_UNLOCK_SELF", "admins cannot unlock their own account")
		return
	}

	// 4. Fetch target user — returns 404 for both missing and soft-deleted users.
	user, err := h.userRepo.FindByID(r.Context(), targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
			return
		}
		h.logger.Error("failed to fetch user for unlock", "target_id", targetID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	// 5. Reject if account is not currently locked — prevents silent no-ops.
	if user.LockedAt == nil {
		writeError(w, http.StatusConflict, "USER_NOT_LOCKED", "user account is not locked")
		return
	}

	// 6. Clear the lock and reset the failed-attempts counter.
	if err := h.userRepo.Unlock(r.Context(), targetID); err != nil {
		h.logger.Error("failed to unlock account", "target_id", targetID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	// 7. Write the audit log (pure computation → persist).
	auditEntry := service.BuildUnlockAuditLog(targetID, admin.ID)
	if err := h.auditRepo.Create(r.Context(), auditEntry); err != nil {
		h.logger.Error("failed to write audit log for unlock", "target_id", targetID, "error", err)
		// Audit failure is logged but does not roll back the unlock — the operation succeeded.
	}

	writeSuccess(w, http.StatusOK, "USER_UNLOCKED", map[string]any{
		"user": map[string]any{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  string(user.Role),
		},
	})
}
