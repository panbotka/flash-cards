package models

import "time"

type Card struct {
	ID        int64      `json:"id" db:"id"`
	Czech     string     `json:"czech" db:"czech"`
	English   string     `json:"english" db:"english"`
	DeletedAt *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
	Suspended bool       `json:"suspended" db:"suspended"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
	Tags      []string   `json:"tags,omitempty"`
	SRSStates []SRSState `json:"srsStates,omitempty"`
}

type SRSState struct {
	ID           int64     `json:"id" db:"id"`
	CardID       int64     `json:"cardId" db:"card_id"`
	Direction    string    `json:"direction" db:"direction"`
	EaseFactor   float64   `json:"easeFactor" db:"ease_factor"`
	IntervalDays float64   `json:"intervalDays" db:"interval_days"`
	Repetitions  int       `json:"repetitions" db:"repetitions"`
	NextReview   time.Time `json:"nextReview" db:"next_review"`
	Status       string    `json:"status" db:"status"`
	LearningStep int       `json:"learningStep" db:"learning_step"`
}

type ReviewEvent struct {
	ID             int64     `json:"id" db:"id"`
	SRSStateID     int64     `json:"srsStateId" db:"srs_state_id"`
	CardID         int64     `json:"cardId" db:"card_id"`
	Direction      string    `json:"direction" db:"direction"`
	Rating         int       `json:"rating" db:"rating"`
	ReviewedAt     time.Time `json:"reviewedAt" db:"reviewed_at"`
	IntervalBefore *float64  `json:"intervalBefore,omitempty" db:"interval_before"`
	IntervalAfter  *float64  `json:"intervalAfter,omitempty" db:"interval_after"`
	EaseBefore     *float64  `json:"easeBefore,omitempty" db:"ease_before"`
	EaseAfter      *float64  `json:"easeAfter,omitempty" db:"ease_after"`
}

// Request/response types

type CreateCardRequest struct {
	Czech   string   `json:"czech" binding:"required"`
	English string   `json:"english" binding:"required"`
	Tags    []string `json:"tags"`
}

type UpdateCardRequest struct {
	Czech   *string   `json:"czech"`
	English *string   `json:"english"`
	Tags    *[]string `json:"tags"`
}

type ReviewRequest struct {
	SRSStateID int64 `json:"srsStateId" binding:"required"`
	Rating     int   `json:"rating" binding:"required,min=1,max=4"`
}

type ImportPreviewRequest struct {
	Content string `json:"content" binding:"required"`
}

type ImportCommitRequest struct {
	Cards []ImportCard `json:"cards" binding:"required"`
	Tags  []string     `json:"tags"`
}

type ImportCard struct {
	Czech   string `json:"czech"`
	English string `json:"english"`
}

type LoginRequest struct {
	Password string `json:"password" binding:"required"`
}

// StudyDoneResponse is returned when all cards are done for the session.
type StudyDoneResponse struct {
	Done         bool `json:"done"`
	NewAvailable int  `json:"newAvailable"`
}

// StudyCardResponse wraps a card and its SRS state for the study view.
type StudyCardResponse struct {
	Card     Card     `json:"card"`
	SRSState SRSState `json:"srsState"`
}

// ReviewResponse is returned after submitting a rating.
type ReviewResponse struct {
	SRSState     SRSState `json:"srsState"`
	NextInterval string   `json:"nextInterval"` // human-readable like "10m", "1d", "3d"
}
