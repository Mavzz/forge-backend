package models

type Streak struct {
	ID             uint `json:"id"`
	UserID         uint `json:"user_id"`
	CurrentStreak  uint `json:"current_streak"`
	LongestStreak  uint `json:"longest_streak"`
	CompletedToday bool `json:"completed_today"`
}
