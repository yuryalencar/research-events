package repository_test

// Spec: specs/backend/admin-users-unlock.yaml
// Rule: "Write AuditLog: entity_type=user, action=unlocked, changed_by_id=admin ID"

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
)

func TestAuditRepository_Create_PersistsAuditLog(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	// Seed a user to act as ChangedBy (foreign key requirement).
	admin := model.User{Name: "Admin", Email: "admin-audit@example.com", Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&admin).Error)

	repo := repository.NewAuditRepository(tx)
	entry := model.AuditLog{
		EntityType:  model.AuditEntityUser,
		EntityID:    42,
		Action:      model.AuditActionUnlocked,
		ChangedByID: admin.ID,
		Diff:        model.JSONB(`{"locked_at":{"before":"2024-01-01","after":null}}`),
	}

	err := repo.Create(context.Background(), entry)
	require.NoError(t, err)

	var saved model.AuditLog
	require.NoError(t, tx.Last(&saved).Error)
	assert.Equal(t, model.AuditEntityUser, saved.EntityType)
	assert.Equal(t, uint(42), saved.EntityID)
	assert.Equal(t, model.AuditActionUnlocked, saved.Action)
	assert.Equal(t, admin.ID, saved.ChangedByID)
}

func TestAuditRepository_Create_PreservesJSONBDiff(t *testing.T) {
	// Rule: Diff field stores JSONB — verify the raw JSON survives a round-trip through the DB.
	tx, rollback := beginTx(t)
	defer rollback()

	admin := model.User{Name: "Admin2", Email: "admin2-audit@example.com", Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&admin).Error)

	repo := repository.NewAuditRepository(tx)
	diff := model.JSONB(`{"failed_login_attempts":{"before":5,"after":0}}`)
	entry := model.AuditLog{
		EntityType:  model.AuditEntityUser,
		EntityID:    99,
		Action:      model.AuditActionUnlocked,
		ChangedByID: admin.ID,
		Diff:        diff,
	}

	require.NoError(t, repo.Create(context.Background(), entry))

	var saved model.AuditLog
	require.NoError(t, tx.Last(&saved).Error)
	assert.JSONEq(t, string(diff), string(saved.Diff))
}
