package model

import (
	"time"

	"gorm.io/gorm"
)

// --- Types ---

// EventStatus represents the review lifecycle of an event.
// Using a named string type (not iota) keeps values readable in DB and JSON —
// the column stores "pending"/"approved"/"rejected" directly, no lookup table needed.
type EventStatus string

const (
	EventStatusPending  EventStatus = "pending"
	EventStatusApproved EventStatus = "approved"
	EventStatusRejected EventStatus = "rejected"
)

// Event represents a research conference or event submitted for review.
//
// Latitude/Longitude are set directly by the submitter via the Leaflet map picker —
// never geocoded automatically. Domain is an intentionally open string (no DB CHECK
// constraint) so new domains can be added without a migration; validation of allowed
// values happens in the service layer.
type Event struct {
	gorm.Model // embeds ID, CreatedAt, UpdatedAt, DeletedAt (soft delete)

	Name       string      `gorm:"not null"`
	Slug       string      `gorm:"uniqueIndex;not null"`
	Country    string      `gorm:"not null"`
	City       string      `gorm:"not null"`
	Latitude   float64     `gorm:"not null"`
	Longitude  float64     `gorm:"not null"`
	StartDate  time.Time   `gorm:"not null"`
	EndDate    time.Time   `gorm:"not null"`
	WebsiteURL string      `gorm:"not null"`
	Domain     string      `gorm:"not null;index"`
	Status     EventStatus `gorm:"not null;default:pending;index"`
	Year       int         `gorm:"not null;index"`

	CreatedByID     uint
	CreatedBy       User `gorm:"foreignKey:CreatedByID"`
	LastUpdatedByID uint
	LastUpdatedBy   User `gorm:"foreignKey:LastUpdatedByID"`

	// Deadlines is a has-many association — GORM auto-populates EventID and
	// inserts these rows in the same Create call when set on a new Event.
	Deadlines []Deadline `gorm:"foreignKey:EventID"`
}
