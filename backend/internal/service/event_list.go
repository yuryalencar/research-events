package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Types ---

// RawListEventsQuery carries the raw (string) query parameters from
// GET /api/v1/events, before any parsing or validation. Keeping this a plain
// string struct lets the handler stay a thin adapter — all parsing/validation
// logic lives in ValidateListEventsQuery, where it can be tested without
// net/http or httptest.
type RawListEventsQuery struct {
	Year               string
	Domain             string
	Country            string
	Status             string
	Tier               string
	FirstDeadlineMonth string
	BBox               string
	Page               string
	PageSize           string
	Pagination         string
}

// ListEventsInput groups the validated, normalized filters for
// EventRepository.ListEvents, with defaults applied per
// specs/backend/events-list.yaml.
type ListEventsInput struct {
	Year               int
	Domain             *string
	Country            *string
	Status             model.EventStatus
	Tier               *string
	FirstDeadlineMonth *int
	BBox               *BBoxInput
	Page               int
	PageSize           int
	PaginationOff      bool
}

// BBoxInput is the parsed ?bbox=minLng,minLat,maxLng,maxLat viewport filter.
type BBoxInput struct {
	MinLng, MinLat, MaxLng, MaxLat float64
}

// --- Package-level validation data ---

// allowedStatuses mirrors the events_status_check CHECK constraint in
// migrations/003_create_events.sql — this is a closed enum.
var allowedStatuses = map[string]model.EventStatus{
	"pending":  model.EventStatusPending,
	"approved": model.EventStatusApproved,
	"rejected": model.EventStatusRejected,
}

// --- Public functions ---

// FP: pure function
// ValidateListEventsQuery depends only on its arguments and performs no I/O —
// currentYear is passed in by the caller (derived from time.Now()) so this
// function never reads the clock itself. Given the same raw query and
// currentYear it always returns the same ListEventsInput (or error), which
// makes every default and validation rule trivially testable: input -> output,
// no mocks, no server, no database.
func ValidateListEventsQuery(raw RawListEventsQuery, currentYear int) (ListEventsInput, error) {
	year := currentYear
	if raw.Year != "" {
		parsed, err := strconv.Atoi(raw.Year)
		if err != nil {
			return ListEventsInput{}, fmt.Errorf("year must be a valid integer")
		}
		year = parsed
	}

	status := model.EventStatusApproved
	if raw.Status != "" {
		parsed, ok := allowedStatuses[raw.Status]
		if !ok {
			return ListEventsInput{}, fmt.Errorf("status %q is not a recognized status", raw.Status)
		}
		status = parsed
	}

	var domain *string
	if raw.Domain != "" {
		if !allowedDomains[raw.Domain] {
			return ListEventsInput{}, fmt.Errorf("domain %q is not a recognized domain", raw.Domain)
		}
		domain = &raw.Domain
	}

	var country *string
	if raw.Country != "" {
		country = &raw.Country
	}

	var tier *string
	if raw.Tier != "" {
		if !allowedTiers[raw.Tier] {
			return ListEventsInput{}, fmt.Errorf("tier %q is not a recognized tier", raw.Tier)
		}
		tier = &raw.Tier
	}

	var firstDeadlineMonth *int
	if raw.FirstDeadlineMonth != "" {
		parsed, err := strconv.Atoi(raw.FirstDeadlineMonth)
		if err != nil || parsed < 1 || parsed > 12 {
			return ListEventsInput{}, fmt.Errorf("first_deadline_month must be an integer between 1 and 12")
		}
		firstDeadlineMonth = &parsed
	}

	bbox, err := parseBBox(raw.BBox)
	if err != nil {
		return ListEventsInput{}, err
	}

	paginationOff, err := parsePagination(raw.Pagination)
	if err != nil {
		return ListEventsInput{}, err
	}

	if paginationOff {
		return ListEventsInput{
			Year:               year,
			Domain:             domain,
			Country:            country,
			Status:             status,
			Tier:               tier,
			FirstDeadlineMonth: firstDeadlineMonth,
			BBox:               bbox,
			Page:               1,
			PaginationOff:      true,
		}, nil
	}

	page := 1
	if raw.Page != "" {
		parsed, err := strconv.Atoi(raw.Page)
		if err != nil || parsed < 1 {
			return ListEventsInput{}, fmt.Errorf("page must be a positive integer")
		}
		page = parsed
	}

	pageSize := 20
	if raw.PageSize != "" {
		parsed, err := strconv.Atoi(raw.PageSize)
		if err != nil || parsed < 1 || parsed > 100 {
			return ListEventsInput{}, fmt.Errorf("page_size must be an integer between 1 and 100")
		}
		pageSize = parsed
	}

	return ListEventsInput{
		Year:               year,
		Domain:             domain,
		Country:            country,
		Status:             status,
		Tier:               tier,
		FirstDeadlineMonth: firstDeadlineMonth,
		BBox:               bbox,
		Page:               page,
		PageSize:           pageSize,
	}, nil
}

// --- Private functions ---

// FP: pure function
// parseBBox depends only on its argument — no I/O. An empty string returns
// (nil, nil), meaning "no bbox filter". A non-empty string must be exactly
// four comma-separated numbers within valid longitude/latitude ranges, with
// min < max on each axis.
func parseBBox(raw string) (*BBoxInput, error) {
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("bbox must have exactly 4 comma-separated values: minLng,minLat,maxLng,maxLat")
	}

	values := make([]float64, 4)
	for i, part := range parts {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("bbox values must be numbers")
		}
		values[i] = parsed
	}

	bbox := BBoxInput{MinLng: values[0], MinLat: values[1], MaxLng: values[2], MaxLat: values[3]}

	if bbox.MinLng < -180 || bbox.MinLng > 180 || bbox.MaxLng < -180 || bbox.MaxLng > 180 {
		return nil, fmt.Errorf("bbox longitude values must be between -180 and 180")
	}
	if bbox.MinLat < -90 || bbox.MinLat > 90 || bbox.MaxLat < -90 || bbox.MaxLat > 90 {
		return nil, fmt.Errorf("bbox latitude values must be between -90 and 90")
	}
	if bbox.MinLng >= bbox.MaxLng {
		return nil, fmt.Errorf("bbox minLng must be less than maxLng")
	}
	if bbox.MinLat >= bbox.MaxLat {
		return nil, fmt.Errorf("bbox minLat must be less than maxLat")
	}

	return &bbox, nil
}

// FP: pure function
// parsePagination depends only on its argument — no I/O. An empty string
// means "on" (the default). Any value other than "on"/"off" is an error.
func parsePagination(raw string) (bool, error) {
	switch raw {
	case "", "on":
		return false, nil
	case "off":
		return true, nil
	default:
		return false, fmt.Errorf("pagination must be 'on' or 'off'")
	}
}
