package postgres

import "time"

type AchievementReference struct {
	ID                 string     `json:"id"`
	StudentID          string     `json:"student_id"`
	MongoAchievementID string     `json:"mongo_achievement_id"`
	Status             string     `json:"status"` // DRAFT, PENDING, VERIFIED, REJECTED
	RejectionNote      *string    `json:"rejection_note,omitempty"`
	VerifiedBy         *string    `json:"verified_by,omitempty"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type AchievementHistory struct {
	ID             string    `json:"id"`
	AchievementID  string    `json:"achievement_id"`
	ChangedBy      string    `json:"changed_by"`
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	Remarks        string    `json:"remarks"`
	CreatedAt      time.Time `json:"created_at"`
}