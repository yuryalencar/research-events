package model

import (
	"time"

	"gorm.io/gorm"
)

// --- Types ---

// DeadlineType identifies which kind of deadline a Deadline row represents.
type DeadlineType string

const (
	DeadlineTypeAbstract     DeadlineType = "abstract"
	DeadlineTypePaper        DeadlineType = "paper"
	DeadlineTypeNotification DeadlineType = "notification"
	DeadlineTypeCameraReady  DeadlineType = "camera_ready"
	DeadlineTypeOther        DeadlineType = "other"
)

// Deadline belongs to an Event and is immutable once created.
//
// Deadlines are never updated in place: a change creates a new row, marks the
// old one is_active=false, and points SupersededByID at the replacement. The UI
// shows only is_active=true deadlines by default, with a "view history" toggle
// revealing the full chain per type.
type Deadline struct {
	gorm.Model

	EventID     uint         `gorm:"not null;index"`
	Type        DeadlineType `gorm:"not null"`
	Description string       `gorm:"not null"`
	Date        time.Time    `gorm:"not null"`

	// Time and Timezone are independently optional — a deadline may have a
	// date only, a date+time, or a date+time+timezone (e.g. "23:59" + "AoE").
	Time     *string `gorm:"type:varchar(5)"`
	Timezone *string `gorm:"type:varchar(50)"`

	IsOptional bool `gorm:"not null;default:false"`
	IsActive   bool `gorm:"not null;default:true"`

	// SupersededByID points at the Deadline that replaced this one, once superseded.
	SupersededByID *uint

	CreatedByID uint
	CreatedBy   User `gorm:"foreignKey:CreatedByID"`
}
