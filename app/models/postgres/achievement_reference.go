package postgres

import "time"

type AchievementReference struct {
	ID                string     `json:"id"`
	StudentID         string     `json:"student_id"`
	MongoAchievementID string    `json:"mongo_achievement_id"`
	Status            string     `json:"status"` // 'draft', 'submitted', 'verified', 'rejected'
	SubmittedAt       *time.Time `json:"submitted_at,omitempty"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
	VerifiedBy        *string    `json:"verified_by,omitempty"`
	RejectionNote     *string    `json:"rejection_note,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
