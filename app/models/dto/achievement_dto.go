package dto

import (
	"pelaporan_prestasi/app/models/mongo"
	"pelaporan_prestasi/app/models/postgres"
)

type AchievementResponse struct {
	RefID         string              `json:"ref_id"`
	Status        string              `json:"status"`
	RejectionNote *string             `json:"rejection_note,omitempty"`
	Content       mongodb.Achievement `json:"content"`
}

func ToAchievementResponse(ref postgres.AchievementReference, data mongodb.Achievement) AchievementResponse {
	return AchievementResponse{
		RefID:         ref.ID,
		Status:        ref.Status,
		RejectionNote: ref.RejectionNote,
		Content:       data,
	}
}