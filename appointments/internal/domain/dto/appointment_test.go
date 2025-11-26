package dto

import (
	"time"
	"github.com/julesChu12/fly/appointments/internal/domain/entity"
)

// ConflictCheckResponse 冲突检查响应 (added for testing)
type ConflictCheckResponse struct {
	HasConflict   bool        `json:"has_conflict"`
	Conflicts     []*entity.Appointment `json:"conflicts,omitempty"`
	ConflictCount int         `json:"conflict_count"`
	Suggestions   []time.Time `json:"suggestions,omitempty"`
}