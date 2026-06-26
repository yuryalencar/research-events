package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/yuryalencar/research-events/internal/middleware"
	"github.com/yuryalencar/research-events/internal/model"
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

// List handles GET /api/v1/admin/users.
// Admin only — returns a paginated, filtered list of users (no password or token fields).
func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
	raw := toRawListUsersQuery(r.URL.Query())

	input, err := service.ValidateListUsersQuery(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	users, total, err := h.userRepo.List(r.Context(), service.ToListUsersFilter(input))
	if err != nil {
		h.logger.Error("failed to list users", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	data := make([]userListItemResponse, 0, len(users))
	for _, u := range users {
		data = append(data, toUserListItemResponse(u))
	}

	writeSuccessWithMeta(w, http.StatusOK, "USERS_LISTED", data, listMeta{Page: input.Page, Total: total})
}

// Register handles POST /api/v1/admin/users.
// Admin only — registers a new admin or moderator with a hashed password.
func (h *AdminUserHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request body.
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// 2. Validate: required fields, allowed role, password complexity.
	if err := service.ValidateRegisterInput(input.Name, input.Email, input.Password, input.Role); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// 3. Check email uniqueness (including soft-deleted rows — they still hold the unique slot).
	exists, err := h.userRepo.ExistsByEmail(r.Context(), input.Email)
	if err != nil {
		h.logger.Error("failed to check email existence", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	if exists {
		writeError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "a user with this email already exists")
		return
	}

	// 4. Hash password — computation only, no I/O.
	hash, err := service.HashPassword(input.Password)
	if err != nil {
		h.logger.Error("failed to hash password", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	// 5. Build user value (pure, no mutation).
	newUser := service.BuildRegisterUser(input.Name, input.Email, hash, model.UserRole(input.Role))

	// 6. Persist.
	created, err := h.userRepo.Create(r.Context(), newUser)
	if err != nil {
		h.logger.Error("failed to create user", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	// 7. Write audit log (pure computation → persist).
	admin, _ := middleware.AuthUserFromContext(r.Context())
	auditEntry := service.BuildRegisterAuditLog(created.ID, admin.ID, created.Name, created.Email, created.Role)
	if err := h.auditRepo.Create(r.Context(), auditEntry); err != nil {
		h.logger.Error("failed to write audit log for user registration", "user_id", created.ID, "error", err)
	}

	writeSuccess(w, http.StatusCreated, "USER_REGISTERED", map[string]any{
		"user": map[string]any{
			"id":         created.ID,
			"name":       created.Name,
			"email":      created.Email,
			"role":       string(created.Role),
			"created_at": created.CreatedAt,
		},
	})
}

// ChangeRole handles PATCH /api/v1/admin/users/{id}/role.
// Admin only — changes a user's role and invalidates their active session.
func (h *AdminUserHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	// 1. Parse target user ID.
	rawID := r.PathValue("id")
	targetID64, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", ":id must be a valid integer")
		return
	}
	targetID := uint(targetID64)

	// 2. Self-change guard — checked before any DB call.
	admin, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	if targetID == admin.ID {
		writeError(w, http.StatusUnprocessableEntity, "CANNOT_CHANGE_OWN_ROLE", "admins cannot change their own role")
		return
	}

	// 3. Parse and validate role from body.
	var input struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	if err := service.ValidateRoleChangeInput(input.Role); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// 4. Fetch target user.
	user, err := h.userRepo.FindByID(r.Context(), targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
			return
		}
		h.logger.Error("failed to fetch user for role change", "target_id", targetID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	// 5. No-op guard — same role is a conflict, not a silent success.
	newRole := model.UserRole(input.Role)
	if user.Role == newRole {
		writeError(w, http.StatusConflict, "ROLE_UNCHANGED", "user already has role '"+string(newRole)+"'")
		return
	}

	// 6. Persist role change.
	if err := h.userRepo.UpdateRole(r.Context(), targetID, newRole); err != nil {
		h.logger.Error("failed to update role", "target_id", targetID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	// 7. Invalidate active session — forces re-login under the new role.
	if err := h.userRepo.ClearTokens(r.Context(), targetID); err != nil {
		h.logger.Error("failed to clear tokens after role change", "target_id", targetID, "error", err)
	}

	// 8. Write audit log (pure computation → persist).
	auditEntry := service.BuildRoleChangedAuditLog(targetID, admin.ID, user.Role, newRole)
	if err := h.auditRepo.Create(r.Context(), auditEntry); err != nil {
		h.logger.Error("failed to write audit log for role change", "target_id", targetID, "error", err)
	}

	writeSuccess(w, http.StatusOK, "USER_ROLE_CHANGED", map[string]any{
		"user": map[string]any{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  string(newRole),
		},
	})
}

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

// --- Private functions ---

// userListItemResponse is the per-user shape returned by GET /api/v1/admin/users.
// Only safe fields are included — password_hash and all token fields are omitted.
type userListItemResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	LockedAt  *time.Time `json:"locked_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// toRawListUsersQuery copies the relevant query string parameters into
// service.RawListUsersQuery. No parsing or validation happens here — that is
// ValidateListUsersQuery's job, so it stays testable without net/http.
func toRawListUsersQuery(q url.Values) service.RawListUsersQuery {
	return service.RawListUsersQuery{
		Roles:          q.Get("roles"),
		Search:         q.Get("search"),
		Locked:         q.Get("locked"),
		IncludeDeleted: q.Get("include_deleted"),
		Page:           q.Get("page"),
		PageSize:       q.Get("page_size"),
		Pagination:     q.Get("pagination"),
	}
}

// toUserListItemResponse maps a model.User to the safe response shape,
// explicitly excluding password_hash and all token fields.
func toUserListItemResponse(u model.User) userListItemResponse {
	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		deletedAt = &u.DeletedAt.Time
	}
	return userListItemResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		LockedAt:  u.LockedAt,
		DeletedAt: deletedAt,
	}
}
