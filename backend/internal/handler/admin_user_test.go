package handler_test

// Spec: specs/backend/admin-users-unlock.yaml
// Rule: "Only admins can unlock — admin cannot unlock themselves"
// Rule: "Write AuditLog with entity_type=user, action=unlocked"

import (
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

