package handler_test

// Spec: specs/backend/admin-users-unlock.yaml
// Spec: specs/backend/admin-users-register.yaml
// Spec: specs/backend/admin-users-change-role.yaml
// Spec: specs/backend/admin-users-list.yaml
// Rule: "Only admins can unlock — admin cannot unlock themselves"
// Rule: "Write AuditLog with entity_type=user, action=unlocked"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/handler"
	"github.com/yuryalencar/research-events/internal/middleware"
	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/repository/mocks"
)

// unlockMux registers the Unlock handler on its real route pattern so that
// r.PathValue("id") is populated by the Go 1.22 ServeMux.
func unlockMux(h *handler.AdminUserHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/admin/users/{id}/unlock", h.Unlock)
	return mux
}

// unlockReq builds a PATCH request with the admin AuthUser injected into context.
func unlockReq(adminID uint, targetID any) *http.Request {
	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/api/v1/admin/users/%v/unlock", targetID),
		nil,
	)
	return req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{
		ID:   adminID,
		Role: "admin",
		Name: "Admin",
	}))
}

func TestAdminUserHandler_Unlock_Success_Returns200(t *testing.T) {
	// Spec: admin-users-unlock.yaml — admin unlocks a different locked user → 200 USER_UNLOCKED
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	adminID := uint(1)
	targetID := uint(2)
	lockedAt := time.Now().Add(-time.Hour)
	target := model.User{
		Model:    gorm.Model{ID: targetID},
		Name:     "Locked User",
		Email:    "locked@example.com",
		Role:     model.UserRoleModerator,
		LockedAt: &lockedAt,
	}

	mockUserRepo.EXPECT().FindByID(gomock.Any(), targetID).Return(target, nil)
	mockUserRepo.EXPECT().Unlock(gomock.Any(), targetID).Return(nil)
	mockAuditRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	unlockMux(h).ServeHTTP(rec, unlockReq(adminID, targetID))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "USER_UNLOCKED", responseCode(t, rec))
}

func TestAdminUserHandler_Unlock_CannotUnlockSelf_Returns422(t *testing.T) {
	// Spec: admin-users-unlock.yaml — admin targeting own ID → 422 CANNOT_UNLOCK_SELF
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)
	// No repo calls expected — self-unlock is rejected before any DB lookup.

	adminID := uint(1)
	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	unlockMux(h).ServeHTTP(rec, unlockReq(adminID, adminID))

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code) // 422
	assert.Equal(t, "CANNOT_UNLOCK_SELF", responseCode(t, rec))
}

func TestAdminUserHandler_Unlock_UserNotFound_Returns404(t *testing.T) {
	// Spec: admin-users-unlock.yaml — target user does not exist → 404 USER_NOT_FOUND
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	mockUserRepo.EXPECT().FindByID(gomock.Any(), uint(99)).Return(model.User{}, repository.ErrNotFound)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	unlockMux(h).ServeHTTP(rec, unlockReq(1, 99))

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "USER_NOT_FOUND", responseCode(t, rec))
}

func TestAdminUserHandler_Unlock_SoftDeletedUser_Returns404(t *testing.T) {
	// Spec: admin-users-unlock.yaml — soft-deleted user → 404 USER_NOT_FOUND (same as missing)
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	// GORM soft-deletes set deleted_at — FindByID already returns ErrNotFound for them.
	mockUserRepo.EXPECT().FindByID(gomock.Any(), uint(5)).Return(model.User{}, repository.ErrNotFound)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	unlockMux(h).ServeHTTP(rec, unlockReq(1, 5))

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "USER_NOT_FOUND", responseCode(t, rec))
}

func TestAdminUserHandler_Unlock_UserNotLocked_Returns409(t *testing.T) {
	// Spec: admin-users-unlock.yaml — target exists but locked_at IS NULL → 409 USER_NOT_LOCKED
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	target := model.User{
		Model:    gorm.Model{ID: 3},
		Role:     model.UserRoleAdmin,
		LockedAt: nil, // not locked
	}
	mockUserRepo.EXPECT().FindByID(gomock.Any(), uint(3)).Return(target, nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	unlockMux(h).ServeHTTP(rec, unlockReq(1, 3))

	assert.Equal(t, http.StatusConflict, rec.Code) // 409
	assert.Equal(t, "USER_NOT_LOCKED", responseCode(t, rec))
}

func TestAdminUserHandler_Unlock_InvalidID_Returns400(t *testing.T) {
	// Spec: admin-users-unlock.yaml border_case ":id is not a valid integer → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)
	// No repo calls expected for invalid ID.

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	unlockMux(h).ServeHTTP(rec, unlockReq(1, "abc"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_Unlock_WritesAuditLogWithCorrectFields(t *testing.T) {
	// Spec: admin-users-unlock.yaml rule "Write AuditLog: entity_type=user, action=unlocked, changed_by_id=admin ID"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	adminID := uint(1)
	targetID := uint(4)
	lockedAt := time.Now().Add(-time.Hour)
	target := model.User{
		Model:    gorm.Model{ID: targetID},
		Role:     model.UserRoleAdmin,
		LockedAt: &lockedAt,
	}

	mockUserRepo.EXPECT().FindByID(gomock.Any(), targetID).Return(target, nil)
	mockUserRepo.EXPECT().Unlock(gomock.Any(), targetID).Return(nil)

	// Verify the audit log has the exact entity type, action, and changed_by_id required by spec.
	mockAuditRepo.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(model.AuditLog{})).
		DoAndReturn(func(_ any, log model.AuditLog) error {
			assert.Equal(t, model.AuditEntityUser, log.EntityType)
			assert.Equal(t, targetID, log.EntityID)
			assert.Equal(t, model.AuditActionUnlocked, log.Action)
			assert.Equal(t, adminID, log.ChangedByID)
			return nil
		})

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	unlockMux(h).ServeHTTP(rec, unlockReq(adminID, targetID))

	require.Equal(t, http.StatusOK, rec.Code)
}

// --- Register helpers ---

func registerMux(h *handler.AdminUserHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/admin/users", h.Register)
	return mux
}

func registerReq(adminID uint, body map[string]any) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{
		ID:   adminID,
		Role: "admin",
		Name: "Admin",
	}))
}

func validRegisterBody() map[string]any {
	return map[string]any{
		"name":     "Jane Doe",
		"email":    "jane@example.com",
		"password": "Secret@123",
		"role":     "moderator",
	}
}

// --- Register tests (Cycle 7) ---

func TestAdminUserHandler_Register_Returns201ForValidModeratorRegistration(t *testing.T) {
	// Spec: admin-users-register.yaml DoD "Returns 201 + user data on valid registration"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	adminID := uint(1)
	hash := "bcrypt-hash"
	created := model.User{Model: gorm.Model{ID: 10}, Name: "Jane Doe", Email: "jane@example.com", Role: model.UserRoleModerator, PasswordHash: &hash}

	mockUserRepo.EXPECT().ExistsByEmail(gomock.Any(), "jane@example.com").Return(false, nil)
	mockUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(created, nil)
	mockAuditRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(adminID, validRegisterBody()))

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "USER_REGISTERED", responseCode(t, rec))
}

func TestAdminUserHandler_Register_Returns201ForValidAdminRegistration(t *testing.T) {
	// Spec: admin-users-register.yaml border_case "Admin registers another admin → 201"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	hash := "bcrypt-hash"
	created := model.User{Model: gorm.Model{ID: 11}, Name: "John Admin", Email: "john@example.com", Role: model.UserRoleAdmin, PasswordHash: &hash}

	mockUserRepo.EXPECT().ExistsByEmail(gomock.Any(), "john@example.com").Return(false, nil)
	mockUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(created, nil)
	mockAuditRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	body := map[string]any{"name": "John Admin", "email": "john@example.com", "password": "Secret@123", "role": "admin"}
	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(1, body))

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "USER_REGISTERED", responseCode(t, rec))
}

func TestAdminUserHandler_Register_Returns400ForMissingName(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	body := validRegisterBody()
	delete(body, "name")
	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(1, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_Register_Returns400ForMissingEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	body := validRegisterBody()
	delete(body, "email")
	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(1, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_Register_Returns400ForMissingPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	body := validRegisterBody()
	delete(body, "password")
	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(1, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_Register_Returns400ForMissingRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	body := validRegisterBody()
	delete(body, "role")
	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(1, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_Register_Returns400ForContributorRole(t *testing.T) {
	// Spec: admin-users-register.yaml border_case "role = contributor → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	body := validRegisterBody()
	body["role"] = "contributor"
	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(1, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_Register_Returns400ForInvalidRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	body := validRegisterBody()
	body["role"] = "superuser"
	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(1, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_Register_Returns400ForWeakPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	body := validRegisterBody()
	body["password"] = "weakpass"
	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(1, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_Register_Returns409WhenEmailAlreadyExists(t *testing.T) {
	// Spec: admin-users-register.yaml DoD "Returns 409 EMAIL_ALREADY_EXISTS for email taken"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	mockUserRepo.EXPECT().ExistsByEmail(gomock.Any(), "jane@example.com").Return(true, nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(1, validRegisterBody()))

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "EMAIL_ALREADY_EXISTS", responseCode(t, rec))
}

func TestAdminUserHandler_Register_WritesAuditLogOnSuccess(t *testing.T) {
	// Spec: admin-users-register.yaml rule "Write AuditLog: entity_type=user, action=created"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	adminID := uint(1)
	hash := "bcrypt-hash"
	created := model.User{Model: gorm.Model{ID: 20}, Name: "Jane Doe", Email: "jane@example.com", Role: model.UserRoleModerator, PasswordHash: &hash}

	mockUserRepo.EXPECT().ExistsByEmail(gomock.Any(), "jane@example.com").Return(false, nil)
	mockUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(created, nil)
	mockAuditRepo.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(model.AuditLog{})).
		DoAndReturn(func(_ any, log model.AuditLog) error {
			assert.Equal(t, model.AuditEntityUser, log.EntityType)
			assert.Equal(t, created.ID, log.EntityID)
			assert.Equal(t, model.AuditActionCreated, log.Action)
			assert.Equal(t, adminID, log.ChangedByID)
			return nil
		})

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	registerMux(h).ServeHTTP(rec, registerReq(adminID, validRegisterBody()))

	require.Equal(t, http.StatusCreated, rec.Code)
}

// --- ChangeRole helpers ---

func changeRoleMux(h *handler.AdminUserHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/admin/users/{id}/role", h.ChangeRole)
	return mux
}

func changeRoleReq(adminID uint, targetID any, role string) *http.Request {
	b, _ := json.Marshal(map[string]string{"role": role})
	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/api/v1/admin/users/%v/role", targetID),
		bytes.NewReader(b),
	)
	req.Header.Set("Content-Type", "application/json")
	return req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{
		ID:   adminID,
		Role: "admin",
		Name: "Admin",
	}))
}

// --- ChangeRole tests (Cycle 8) ---

func TestAdminUserHandler_ChangeRole_Returns422WhenAdminTargetsOwnID(t *testing.T) {
	// Spec: admin-users-change-role.yaml — 422 CANNOT_CHANGE_OWN_ROLE checked before any DB call
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	adminID := uint(1)
	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	changeRoleMux(h).ServeHTTP(rec, changeRoleReq(adminID, adminID, "moderator"))

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Equal(t, "CANNOT_CHANGE_OWN_ROLE", responseCode(t, rec))
}

func TestAdminUserHandler_ChangeRole_Returns400ForNonIntegerID(t *testing.T) {
	// Spec: admin-users-change-role.yaml border_case "Non-integer :id → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	changeRoleMux(h).ServeHTTP(rec, changeRoleReq(1, "abc", "moderator"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_ChangeRole_Returns400ForInvalidRole(t *testing.T) {
	// Spec: admin-users-change-role.yaml border_case "Invalid role value in body → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	changeRoleMux(h).ServeHTTP(rec, changeRoleReq(1, 2, "superuser"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_ChangeRole_Returns404WhenUserNotFound(t *testing.T) {
	// Spec: admin-users-change-role.yaml — 404 USER_NOT_FOUND for missing or soft-deleted
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	mockUserRepo.EXPECT().FindByID(gomock.Any(), uint(99)).Return(model.User{}, repository.ErrNotFound)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	changeRoleMux(h).ServeHTTP(rec, changeRoleReq(1, 99, "moderator"))

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "USER_NOT_FOUND", responseCode(t, rec))
}

func TestAdminUserHandler_ChangeRole_Returns409WhenRoleUnchanged(t *testing.T) {
	// Spec: admin-users-change-role.yaml — 409 ROLE_UNCHANGED prevents silent no-ops
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	target := model.User{Model: gorm.Model{ID: 2}, Role: model.UserRoleModerator}
	mockUserRepo.EXPECT().FindByID(gomock.Any(), uint(2)).Return(target, nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	changeRoleMux(h).ServeHTTP(rec, changeRoleReq(1, 2, "moderator"))

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "ROLE_UNCHANGED", responseCode(t, rec))
}

func TestAdminUserHandler_ChangeRole_Returns200AndClearsTokensOnSuccess(t *testing.T) {
	// Spec: admin-users-change-role.yaml DoD "access_token_jti and refresh_token_hash cleared in DB"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	target := model.User{Model: gorm.Model{ID: 2}, Name: "Bob", Email: "bob@example.com", Role: model.UserRoleContributor}
	mockUserRepo.EXPECT().FindByID(gomock.Any(), uint(2)).Return(target, nil)
	mockUserRepo.EXPECT().UpdateRole(gomock.Any(), uint(2), model.UserRoleModerator).Return(nil)
	mockUserRepo.EXPECT().ClearTokens(gomock.Any(), uint(2)).Return(nil)
	mockAuditRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	changeRoleMux(h).ServeHTTP(rec, changeRoleReq(1, 2, "moderator"))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "USER_ROLE_CHANGED", responseCode(t, rec))
}

func TestAdminUserHandler_ChangeRole_WritesAuditLogWithRoleDiff(t *testing.T) {
	// Spec: admin-users-change-role.yaml rule "Write AuditLog: action=role_changed, diff={role:{before,after}}"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	adminID := uint(1)
	target := model.User{Model: gorm.Model{ID: 2}, Role: model.UserRoleContributor}
	mockUserRepo.EXPECT().FindByID(gomock.Any(), uint(2)).Return(target, nil)
	mockUserRepo.EXPECT().UpdateRole(gomock.Any(), uint(2), model.UserRoleModerator).Return(nil)
	mockUserRepo.EXPECT().ClearTokens(gomock.Any(), uint(2)).Return(nil)
	mockAuditRepo.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(model.AuditLog{})).
		DoAndReturn(func(_ any, log model.AuditLog) error {
			assert.Equal(t, model.AuditEntityUser, log.EntityType)
			assert.Equal(t, uint(2), log.EntityID)
			assert.Equal(t, model.AuditActionRoleChanged, log.Action)
			assert.Equal(t, adminID, log.ChangedByID)
			diffStr := string(log.Diff)
			assert.Contains(t, diffStr, "contributor")
			assert.Contains(t, diffStr, "moderator")
			return nil
		})

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	changeRoleMux(h).ServeHTTP(rec, changeRoleReq(adminID, 2, "moderator"))

	require.Equal(t, http.StatusOK, rec.Code)
}

// --- List helpers ---

func listUsersMux(h *handler.AdminUserHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/admin/users", h.List)
	return mux
}

func listUsersReq(role, query string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?"+query, nil)
	return req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{
		ID:   1,
		Role: role,
		Name: "Caller",
	}))
}

// --- List tests (Cycle 9) ---

func TestAdminUserHandler_List_Returns200WithDataAndMeta(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "200 + paginated list with correct meta for valid admin request"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	users := []model.User{
		{Model: gorm.Model{ID: 1}, Name: "Alice", Email: "alice@example.com", Role: model.UserRoleAdmin},
		{Model: gorm.Model{ID: 2}, Name: "Bob", Email: "bob@example.com", Role: model.UserRoleModerator},
	}
	mockUserRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(users, int64(2), nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	listUsersMux(h).ServeHTTP(rec, listUsersReq("admin", ""))

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "USERS_LISTED", body["code"])
	meta := body["meta"].(map[string]any)
	assert.Equal(t, float64(2), meta["total"])
	assert.Equal(t, float64(1), meta["page"])
}

func TestAdminUserHandler_List_EmptyResult_Returns200NotError(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "200 + empty data array (not 404) when no users match"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	mockUserRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return([]model.User{}, int64(0), nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	listUsersMux(h).ServeHTTP(rec, listUsersReq("admin", "roles=admin"))

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "USERS_LISTED", body["code"])
	data := body["data"].([]any)
	assert.Empty(t, data)
}

func TestAdminUserHandler_List_InvalidRole_Returns400(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "400 when roles contains an unknown value"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)
	// No repo call expected — validation fails before any DB access.

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	listUsersMux(h).ServeHTTP(rec, listUsersReq("admin", "roles=superuser"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_List_InvalidPage_Returns400(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "400 when page is not a positive integer"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	listUsersMux(h).ServeHTTP(rec, listUsersReq("admin", "page=0"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_List_PageSizeExceedsMax_Returns400(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "400 when page_size exceeds 100"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	listUsersMux(h).ServeHTTP(rec, listUsersReq("admin", "page_size=101"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminUserHandler_List_NeverIncludesPasswordHash(t *testing.T) {
	// Spec: admin-users-list.yaml rule "Response never includes password_hash or any token field"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	secret := "bcrypt-hash-secret"
	users := []model.User{
		{Model: gorm.Model{ID: 1}, Name: "Alice", Email: "alice@example.com", Role: model.UserRoleAdmin, PasswordHash: &secret},
	}
	mockUserRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(users, int64(1), nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	listUsersMux(h).ServeHTTP(rec, listUsersReq("admin", ""))

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.NotContains(t, body, "password_hash", "password_hash must never appear in the response")
	assert.NotContains(t, body, "bcrypt-hash-secret", "plaintext hash value must never appear in the response")
	assert.NotContains(t, body, "access_token", "token fields must never appear in the response")
	assert.NotContains(t, body, "refresh_token", "token fields must never appear in the response")
}

func TestAdminUserHandler_List_ResponseIncludesLockedAt(t *testing.T) {
	// Spec: admin-users-list.yaml — locked_at field present in each user object (null when not locked)
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	lockedAt := time.Now()
	users := []model.User{
		{Model: gorm.Model{ID: 1}, Name: "Locked", Email: "locked@example.com", Role: model.UserRoleAdmin, LockedAt: &lockedAt},
		{Model: gorm.Model{ID: 2}, Name: "Active", Email: "active@example.com", Role: model.UserRoleAdmin},
	}
	mockUserRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(users, int64(2), nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	listUsersMux(h).ServeHTTP(rec, listUsersReq("admin", ""))

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))

	data := body["data"].([]any)
	require.Len(t, data, 2)

	lockedUser := data[0].(map[string]any)
	activeUser := data[1].(map[string]any)
	assert.NotNil(t, lockedUser["locked_at"])
	assert.Nil(t, activeUser["locked_at"])
}

func TestAdminUserHandler_List_PassesRoleFilterToRepository(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "200 + only admins when roles=admin"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	mockUserRepo.EXPECT().
		List(gomock.Any(), gomock.AssignableToTypeOf(repository.ListUsersFilter{})).
		DoAndReturn(func(_ any, f repository.ListUsersFilter) ([]model.User, int64, error) {
			assert.Equal(t, []model.UserRole{model.UserRoleAdmin}, f.Roles)
			return []model.User{}, int64(0), nil
		})

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	listUsersMux(h).ServeHTTP(rec, listUsersReq("admin", "roles=admin"))

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminUserHandler_List_ResponseIncludesDeletedAt(t *testing.T) {
	// Spec: admin-users-list.yaml — deleted_at is set for soft-deleted users, null for active ones
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	deletedTime := time.Now().Add(-24 * time.Hour)
	users := []model.User{
		{
			Model: gorm.Model{ID: 1, DeletedAt: gorm.DeletedAt{Time: deletedTime, Valid: true}},
			Name:  "Deleted User", Email: "deleted@example.com", Role: model.UserRoleAdmin,
		},
		{
			Model: gorm.Model{ID: 2},
			Name:  "Active User", Email: "active@example.com", Role: model.UserRoleAdmin,
		},
	}
	mockUserRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(users, int64(2), nil)

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	listUsersMux(h).ServeHTTP(rec, listUsersReq("admin", "include_deleted=true"))

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))

	data := body["data"].([]any)
	require.Len(t, data, 2)

	deletedUser := data[0].(map[string]any)
	activeUser := data[1].(map[string]any)
	assert.NotNil(t, deletedUser["deleted_at"], "deleted user must have deleted_at set")
	assert.Nil(t, activeUser["deleted_at"], "active user must have deleted_at as null")
}

func TestAdminUserHandler_List_PassesSearchFilterToRepository(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "200 + matching users when search=<partial name or email>"
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

	mockUserRepo.EXPECT().
		List(gomock.Any(), gomock.AssignableToTypeOf(repository.ListUsersFilter{})).
		DoAndReturn(func(_ any, f repository.ListUsersFilter) ([]model.User, int64, error) {
			assert.Equal(t, "alice", f.Search)
			return []model.User{}, int64(0), nil
		})

	h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
	rec := httptest.NewRecorder()
	listUsersMux(h).ServeHTTP(rec, listUsersReq("admin", "search=alice"))

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminUserHandler_List_Returns200ForAllValidRoleTransitions(t *testing.T) {
	// Spec: admin-users-change-role.yaml border_cases — all 6 allowed transitions
	transitions := []struct {
		from model.UserRole
		to   model.UserRole
	}{
		{model.UserRoleContributor, model.UserRoleModerator},
		{model.UserRoleContributor, model.UserRoleAdmin},
		{model.UserRoleModerator, model.UserRoleContributor},
		{model.UserRoleModerator, model.UserRoleAdmin},
		{model.UserRoleAdmin, model.UserRoleModerator},
		{model.UserRoleAdmin, model.UserRoleContributor},
	}

	for _, tc := range transitions {
		t.Run(fmt.Sprintf("%s_to_%s", tc.from, tc.to), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockUserRepo := mocks.NewMockUserRepository(ctrl)
			mockAuditRepo := mocks.NewMockAuditRepository(ctrl)

			target := model.User{Model: gorm.Model{ID: 2}, Role: tc.from}
			mockUserRepo.EXPECT().FindByID(gomock.Any(), uint(2)).Return(target, nil)
			mockUserRepo.EXPECT().UpdateRole(gomock.Any(), uint(2), tc.to).Return(nil)
			mockUserRepo.EXPECT().ClearTokens(gomock.Any(), uint(2)).Return(nil)
			mockAuditRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

			h := handler.NewAdminUserHandler(mockUserRepo, mockAuditRepo, testLogger)
			rec := httptest.NewRecorder()
			changeRoleMux(h).ServeHTTP(rec, changeRoleReq(1, 2, string(tc.to)))

			assert.Equal(t, http.StatusOK, rec.Code, "transition %s→%s", tc.from, tc.to)
		})
	}
}

