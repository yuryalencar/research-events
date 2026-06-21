package handler_test

// Spec: specs/backend/users-update-password.yaml

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/handler"
	"github.com/yuryalencar/research-events/internal/middleware"
	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository/mocks"
)

// --- Helpers ---

var userTestLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

// updatePasswordMux registers UpdatePassword on the real route so handler logic
// can be exercised through the Go 1.22 ServeMux (no path params needed here,
// but consistency with other handler tests makes the setup familiar).
func updatePasswordMux(h *handler.UserHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/users/me/password", h.UpdatePassword)
	return mux
}

// mustHashPassword produces a bcrypt hash at cost 4 (fastest allowed) for use in
// test fixtures only — cost 4 keeps tests fast without compromising production security.
func mustHashPassword(t *testing.T, plain string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 4)
	require.NoError(t, err)
	return string(hash)
}

// updatePasswordRequest builds the JSON body for the UpdatePassword endpoint.
func updatePasswordRequest(current, newPw, confirm string) *bytes.Buffer {
	body, _ := json.Marshal(map[string]string{
		"current_password":          current,
		"new_password":              newPw,
		"new_password_confirmation": confirm,
	})
	return bytes.NewBuffer(body)
}

// --- Tests ---

// Spec DoD: "Returns 200 PASSWORD_UPDATED on valid request with correct current
// password and valid new password"
func TestUserHandler_UpdatePassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currentPlain := "OldPass@1"
	currentHash := mustHashPassword(t, currentPlain)

	user := model.User{
		Model:        gorm.Model{ID: 7},
		PasswordHash: &currentHash,
	}

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRepo.EXPECT().FindByID(gomock.Any(), uint(7)).Return(user, nil)
	mockRepo.EXPECT().UpdatePassword(gomock.Any(), uint(7), gomock.Any()).Return(nil)

	h := handler.NewUserHandler(mockRepo, userTestLogger)
	srv := updatePasswordMux(h)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/password",
		updatePasswordRequest(currentPlain, "NewPass@2", "NewPass@2"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{ID: 7}))

	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "PASSWORD_UPDATED", resp["code"])
}

// Spec DoD: "Returns 401 when request has no JWT cookie" — in the handler
// this means AuthUserFromContext returns false (RequireAuth not in chain).
func TestUserHandler_UpdatePassword_NoAuthUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	// No repo calls expected — handler must reject before any DB lookup.

	h := handler.NewUserHandler(mockRepo, userTestLogger)
	srv := updatePasswordMux(h)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/password",
		updatePasswordRequest("OldPass@1", "NewPass@2", "NewPass@2"))
	req.Header.Set("Content-Type", "application/json")
	// Intentionally no middleware.WithAuthUser — simulates missing JWT.

	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "TOKEN_MISSING", resp["code"])
}

// Spec DoD: "Returns 400 VALIDATION_ERROR on malformed JSON body"
func TestUserHandler_UpdatePassword_MalformedJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	h := handler.NewUserHandler(mockRepo, userTestLogger)
	srv := updatePasswordMux(h)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/password",
		strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{ID: 7}))

	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp["code"])
}

// Spec DoD: "Returns 400 VALIDATION_ERROR when any required field is missing or empty"
func TestUserHandler_UpdatePassword_MissingFields(t *testing.T) {
	cases := []struct {
		name    string
		current string
		newPw   string
		confirm string
	}{
		{"missing current_password", "", "NewPass@2", "NewPass@2"},
		{"missing new_password", "OldPass@1", "", ""},
		{"missing new_password_confirmation", "OldPass@1", "NewPass@2", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockUserRepository(ctrl)
			// No repo calls expected.

			h := handler.NewUserHandler(mockRepo, userTestLogger)
			srv := updatePasswordMux(h)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/password",
				updatePasswordRequest(tc.current, tc.newPw, tc.confirm))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{ID: 7}))

			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)

			var resp map[string]any
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			assert.Equal(t, "VALIDATION_ERROR", resp["code"])
		})
	}
}

// Spec DoD: "Returns 400 VALIDATION_ERROR when new_password and
// new_password_confirmation do not match"
func TestUserHandler_UpdatePassword_PasswordsMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	h := handler.NewUserHandler(mockRepo, userTestLogger)
	srv := updatePasswordMux(h)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/password",
		updatePasswordRequest("OldPass@1", "NewPass@2", "Different@3"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{ID: 7}))

	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp["code"])
}

// Spec DoD: complexity failures — too short, no upper, no lower, no special.
func TestUserHandler_UpdatePassword_ComplexityFailures(t *testing.T) {
	cases := []struct {
		name  string
		newPw string
	}{
		{"too short", "Ab@1"},
		{"no uppercase", "newpass@1"},
		{"no lowercase", "NEWPASS@1"},
		{"no special character", "NewPass12"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockUserRepository(ctrl)
			// No repo calls expected — complexity is validated before DB lookup.

			h := handler.NewUserHandler(mockRepo, userTestLogger)
			srv := updatePasswordMux(h)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/password",
				updatePasswordRequest("OldPass@1", tc.newPw, tc.newPw))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{ID: 7}))

			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)

			var resp map[string]any
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			assert.Equal(t, "VALIDATION_ERROR", resp["code"])
		})
	}
}

// Spec DoD: "Returns 400 INVALID_CURRENT_PASSWORD when current_password is wrong"
func TestUserHandler_UpdatePassword_WrongCurrentPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	correctHash := mustHashPassword(t, "CorrectPass@1")
	user := model.User{
		Model:        gorm.Model{ID: 7},
		PasswordHash: &correctHash,
	}

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRepo.EXPECT().FindByID(gomock.Any(), uint(7)).Return(user, nil)
	// UpdatePassword must NOT be called when the current password is wrong.

	h := handler.NewUserHandler(mockRepo, userTestLogger)
	srv := updatePasswordMux(h)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/password",
		updatePasswordRequest("WrongPass@9", "NewPass@2", "NewPass@2"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{ID: 7}))

	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "INVALID_CURRENT_PASSWORD", resp["code"])
}
