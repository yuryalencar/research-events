package model

import (
	"database/sql/driver"
	"fmt"

	"gorm.io/gorm"
)

// --- Types ---

// AuditEntityType identifies which table an AuditLog row describes.
type AuditEntityType string

const (
	AuditEntityEvent    AuditEntityType = "event"
	AuditEntityDeadline AuditEntityType = "deadline"
	AuditEntityUser     AuditEntityType = "user" // extended for account unlock
)

// AuditAction describes what happened to the entity.
type AuditAction string

const (
	AuditActionCreated             AuditAction = "created"
	AuditActionUpdated             AuditAction = "updated"
	AuditActionApproved            AuditAction = "approved"
	AuditActionRejected            AuditAction = "rejected"
	AuditActionDeadlineAdded       AuditAction = "deadline_added"
	AuditActionDeadlineSuperseded  AuditAction = "deadline_superseded"
	AuditActionUnlocked            AuditAction = "unlocked"
	AuditActionBatchDeadlinesAdded AuditAction = "batch_deadlines_added"
	AuditActionDeadlineCancelled   AuditAction = "deadline_cancelled"
	AuditActionRoleChanged         AuditAction = "role_changed"
	AuditActionPasswordChanged     AuditAction = "password_changed"
)

// JSONB is a byte slice that maps to a Postgres jsonb column.
//
// Go's database/sql layer communicates with drivers through two interfaces:
//   - driver.Valuer: called when writing — converts Go value → DB wire value
//   - sql.Scanner:   called when reading — converts DB wire value → Go value
//
// GORM does not know how to store a raw []byte as jsonb by default.
// Implementing these two interfaces teaches the driver how to handle the conversion,
// so we can write `gorm:"type:jsonb"` and everything works transparently.
type JSONB []byte

// Value implements driver.Valuer — called by the database driver on every INSERT/UPDATE.
// We return the bytes as a string; the Postgres driver sends it to the jsonb column.
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// Scan implements sql.Scanner — called by the database driver on every SELECT.
// pgx v5 returns jsonb columns as []byte; we copy the bytes into the receiver.
func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = JSONB(v)
	default:
		return fmt.Errorf("JSONB.Scan: unsupported source type %T", value)
	}
	return nil
}

// AuditLog is the immutable change history for events, deadlines, and users.
// Every state change in the system must write one row here.
// The Diff field stores a JSONB object of {"field": {"before": X, "after": Y}} pairs.
type AuditLog struct {
	gorm.Model

	EntityType  AuditEntityType `gorm:"not null"`
	EntityID    uint            `gorm:"not null;index"`
	Action      AuditAction     `gorm:"not null"`
	ChangedByID uint            `gorm:"not null"`
	ChangedBy   User            `gorm:"foreignKey:ChangedByID"`
	Diff        JSONB           `gorm:"type:jsonb"`
	Reason      *string
}
