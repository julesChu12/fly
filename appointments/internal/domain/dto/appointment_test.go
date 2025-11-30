package dto

import (
	"github.com/julesChu12/fly/appointments/internal/domain/appointment"
	"time"
)

// ConflictCheckResponse 冲突检查响应 (added for testing)
type ConflictCheckResponse struct {
	HasConflict   bool                       `json:"has_conflict"`
	Conflicts     []*appointment.Appointment `json:"conflicts,omitempty"`
	ConflictCount int                        `json:"conflict_count"`
	Suggestions   []time.Time                `json:"suggestions,omitempty"`
}
