package srs

import (
	"fmt"
	"math"
	"time"

	"github.com/pi/flash-cards/internal/models"
)

// Learning steps in minutes.
var LearningSteps = []float64{1, 10}

const (
	GraduatingInterval     = 1.0   // days
	EasyGraduatingInterval = 4.0   // days
	MinEaseFactor          = 1.3
	MaxInterval            = 365.0 // days
)

// Rating constants.
const (
	Again = 1
	Hard  = 2
	Good  = 3
	Easy  = 4
)

// ProcessReview applies a review rating to an SRS state and returns a new state.
// The input state is not mutated.
func ProcessReview(state *models.SRSState, rating int) models.SRSState {
	now := time.Now()
	next := *state // copy

	switch state.Status {
	case "new", "learning":
		processLearning(&next, rating, now)
	case "review":
		processReview(&next, rating, now)
	}

	return next
}

// processLearning handles cards in "new" or "learning" status.
func processLearning(s *models.SRSState, rating int, now time.Time) {
	switch rating {
	case Again:
		s.LearningStep = 0
		s.EaseFactor = math.Max(s.EaseFactor-0.20, MinEaseFactor)
		s.Status = "learning"
		s.NextReview = now.Add(time.Duration(LearningSteps[0] * float64(time.Minute)))

	case Hard:
		// Stay on current step.
		s.EaseFactor = math.Max(s.EaseFactor-0.15, MinEaseFactor)
		s.Status = "learning"
		step := clampStep(s.LearningStep)
		s.NextReview = now.Add(time.Duration(LearningSteps[step] * float64(time.Minute)))

	case Good:
		nextStep := s.LearningStep + 1
		if nextStep >= len(LearningSteps) {
			// Graduate the card.
			s.Status = "review"
			s.IntervalDays = GraduatingInterval
			s.Repetitions++
			s.LearningStep = 0
			s.NextReview = now.Add(daysToDuration(GraduatingInterval))
		} else {
			s.LearningStep = nextStep
			s.Status = "learning"
			s.NextReview = now.Add(time.Duration(LearningSteps[nextStep] * float64(time.Minute)))
		}

	case Easy:
		s.Status = "review"
		s.IntervalDays = EasyGraduatingInterval
		s.EaseFactor += 0.15
		s.Repetitions++
		s.LearningStep = 0
		s.NextReview = now.Add(daysToDuration(EasyGraduatingInterval))
	}
}

// processReview handles cards in "review" status.
func processReview(s *models.SRSState, rating int, now time.Time) {
	oldInterval := s.IntervalDays

	switch rating {
	case Again:
		s.LearningStep = 0
		s.EaseFactor = math.Max(s.EaseFactor-0.20, MinEaseFactor)
		s.Status = "learning"
		s.NextReview = now.Add(time.Duration(LearningSteps[0] * float64(time.Minute)))
		// IntervalDays is preserved so it can be restored after re-learning.
		return

	case Hard:
		newInterval := oldInterval * 1.2
		s.IntervalDays = clampInterval(newInterval)
		s.EaseFactor = math.Max(s.EaseFactor-0.15, MinEaseFactor)

	case Good:
		newInterval := oldInterval * s.EaseFactor
		s.IntervalDays = clampInterval(newInterval)
		s.Repetitions++

	case Easy:
		newInterval := oldInterval * s.EaseFactor * 1.3
		s.IntervalDays = clampInterval(newInterval)
		s.EaseFactor += 0.15
		s.Repetitions++
	}

	s.NextReview = now.Add(daysToDuration(s.IntervalDays))
}

// FormatInterval returns a human-readable string for the interval that would
// result from applying the given rating to the state. Used for button labels.
func FormatInterval(state *models.SRSState, rating int) string {
	result := ProcessReview(state, rating)

	switch result.Status {
	case "learning":
		step := clampStep(result.LearningStep)
		minutes := LearningSteps[step]
		return formatMinutes(minutes)
	case "review":
		return formatDays(result.IntervalDays)
	}

	return ""
}

// clampInterval ensures the interval is between 1 day and MaxInterval.
func clampInterval(days float64) float64 {
	if days < 1 {
		days = 1
	}
	if days > MaxInterval {
		days = MaxInterval
	}
	return math.Round(days*10) / 10 // round to 1 decimal place
}

// clampStep ensures the learning step index is within bounds.
func clampStep(step int) int {
	if step < 0 {
		return 0
	}
	if step >= len(LearningSteps) {
		return len(LearningSteps) - 1
	}
	return step
}

// daysToDuration converts a number of days to a time.Duration.
func daysToDuration(days float64) time.Duration {
	return time.Duration(days * 24 * float64(time.Hour))
}

// formatMinutes renders a minute count as a compact string.
func formatMinutes(m float64) string {
	if m < 60 {
		return fmt.Sprintf("%dm", int(m))
	}
	hours := m / 60
	if hours < 24 {
		return fmt.Sprintf("%dh", int(hours))
	}
	return fmt.Sprintf("%dd", int(hours/24))
}

// formatDays renders a day count as a compact human-readable string.
func formatDays(d float64) string {
	days := math.Round(d)
	if days < 1 {
		days = 1
	}
	if days == 1 {
		return "1d"
	}
	if days < 7 {
		return fmt.Sprintf("%dd", int(days))
	}
	if days < 30 {
		weeks := int(math.Round(days / 7))
		if weeks < 1 {
			weeks = 1
		}
		return fmt.Sprintf("%dw", weeks)
	}
	if days < 365 {
		months := int(math.Round(days / 30))
		if months < 1 {
			months = 1
		}
		return fmt.Sprintf("%dmo", months)
	}
	years := days / 365
	return fmt.Sprintf("%.1fy", years)
}
