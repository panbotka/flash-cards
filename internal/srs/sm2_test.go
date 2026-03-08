package srs

import (
	"testing"
	"time"

	"github.com/pi/flash-cards/internal/models"
)

func newState(status string, easeFactor float64, intervalDays float64, step int, reps int) *models.SRSState {
	return &models.SRSState{
		Status:       status,
		EaseFactor:   easeFactor,
		IntervalDays: intervalDays,
		LearningStep: step,
		Repetitions:  reps,
		NextReview:   time.Now(),
	}
}

func TestProcessReview_NewCard(t *testing.T) {
	tests := []struct {
		name           string
		rating         int
		wantStatus     string
		wantStep       int
		wantEase       float64
		wantInterval   float64
		wantRepsChange int // 0 = unchanged, 1 = incremented
	}{
		{
			name:         "Again resets to step 0, decreases ease",
			rating:       Again,
			wantStatus:   "learning",
			wantStep:     0,
			wantEase:     2.30,
			wantInterval: 0,
		},
		{
			name:       "Hard stays on current step, decreases ease",
			rating:     Hard,
			wantStatus: "learning",
			wantStep:   0,
			wantEase:   2.35,
		},
		{
			name:       "Good advances to step 1",
			rating:     Good,
			wantStatus: "learning",
			wantStep:   1,
			wantEase:   2.5,
		},
		{
			name:           "Easy graduates immediately",
			rating:         Easy,
			wantStatus:     "review",
			wantStep:       0,
			wantEase:       2.65,
			wantInterval:   EasyGraduatingInterval,
			wantRepsChange: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newState("new", 2.5, 0, 0, 0)
			result := ProcessReview(state, tt.rating)

			if result.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", result.Status, tt.wantStatus)
			}
			if result.LearningStep != tt.wantStep {
				t.Errorf("learningStep = %d, want %d", result.LearningStep, tt.wantStep)
			}
			if !approxEqual(result.EaseFactor, tt.wantEase) {
				t.Errorf("easeFactor = %.2f, want %.2f", result.EaseFactor, tt.wantEase)
			}
			if tt.wantInterval > 0 && !approxEqual(result.IntervalDays, tt.wantInterval) {
				t.Errorf("intervalDays = %.1f, want %.1f", result.IntervalDays, tt.wantInterval)
			}
			if result.Repetitions != tt.wantRepsChange {
				t.Errorf("repetitions = %d, want %d", result.Repetitions, tt.wantRepsChange)
			}
			// Verify input was not mutated.
			if state.Status != "new" {
				t.Error("original state was mutated")
			}
		})
	}
}

func TestProcessReview_LearningCard(t *testing.T) {
	tests := []struct {
		name       string
		step       int
		rating     int
		wantStatus string
		wantStep   int
	}{
		{
			name:       "Again resets to step 0",
			step:       1,
			rating:     Again,
			wantStatus: "learning",
			wantStep:   0,
		},
		{
			name:       "Hard stays on current step",
			step:       1,
			rating:     Hard,
			wantStatus: "learning",
			wantStep:   1,
		},
		{
			name:       "Good on step 0 advances to step 1",
			step:       0,
			rating:     Good,
			wantStatus: "learning",
			wantStep:   1,
		},
		{
			name:       "Good on last step graduates",
			step:       1,
			rating:     Good,
			wantStatus: "review",
			wantStep:   0,
		},
		{
			name:       "Easy graduates immediately from any step",
			step:       0,
			rating:     Easy,
			wantStatus: "review",
			wantStep:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newState("learning", 2.5, 0, tt.step, 0)
			result := ProcessReview(state, tt.rating)

			if result.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", result.Status, tt.wantStatus)
			}
			if result.LearningStep != tt.wantStep {
				t.Errorf("learningStep = %d, want %d", result.LearningStep, tt.wantStep)
			}
		})
	}
}

func TestProcessReview_LearningGraduation(t *testing.T) {
	state := newState("learning", 2.5, 0, 1, 0)
	result := ProcessReview(state, Good)

	if result.Status != "review" {
		t.Fatalf("status = %q, want review", result.Status)
	}
	if !approxEqual(result.IntervalDays, GraduatingInterval) {
		t.Errorf("intervalDays = %.1f, want %.1f", result.IntervalDays, GraduatingInterval)
	}
	if result.Repetitions != 1 {
		t.Errorf("repetitions = %d, want 1", result.Repetitions)
	}
}

func TestProcessReview_ReviewCard(t *testing.T) {
	tests := []struct {
		name         string
		ease         float64
		interval     float64
		rating       int
		wantStatus   string
		wantMinInterval float64
		wantMaxInterval float64
		wantEaseMin  float64
		wantEaseMax  float64
	}{
		{
			name:         "Again lapses to learning",
			ease:         2.5,
			interval:     10,
			rating:       Again,
			wantStatus:   "learning",
			wantEaseMin:  2.30,
			wantEaseMax:  2.30,
		},
		{
			name:            "Hard increases interval by 1.2x, decreases ease",
			ease:            2.5,
			interval:        10,
			rating:          Hard,
			wantStatus:      "review",
			wantMinInterval: 12.0,
			wantMaxInterval: 12.0,
			wantEaseMin:     2.35,
			wantEaseMax:     2.35,
		},
		{
			name:            "Good increases interval by ease factor",
			ease:            2.5,
			interval:        10,
			rating:          Good,
			wantStatus:      "review",
			wantMinInterval: 25.0,
			wantMaxInterval: 25.0,
			wantEaseMin:     2.5,
			wantEaseMax:     2.5,
		},
		{
			name:            "Easy increases interval by ease*1.3, increases ease",
			ease:            2.5,
			interval:        10,
			rating:          Easy,
			wantStatus:      "review",
			wantMinInterval: 32.5,
			wantMaxInterval: 32.5,
			wantEaseMin:     2.65,
			wantEaseMax:     2.65,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newState("review", tt.ease, tt.interval, 0, 3)
			result := ProcessReview(state, tt.rating)

			if result.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", result.Status, tt.wantStatus)
			}
			if !approxEqual(result.EaseFactor, tt.wantEaseMin) {
				t.Errorf("easeFactor = %.2f, want %.2f", result.EaseFactor, tt.wantEaseMin)
			}
			if tt.wantMinInterval > 0 {
				if result.IntervalDays < tt.wantMinInterval-0.1 || result.IntervalDays > tt.wantMaxInterval+0.1 {
					t.Errorf("intervalDays = %.1f, want [%.1f, %.1f]", result.IntervalDays, tt.wantMinInterval, tt.wantMaxInterval)
				}
			}
		})
	}
}

func TestProcessReview_EaseFactorClamping(t *testing.T) {
	// Ease factor should never go below MinEaseFactor (1.3).
	state := newState("review", 1.35, 10, 0, 3)
	result := ProcessReview(state, Again)

	if result.EaseFactor < MinEaseFactor {
		t.Errorf("easeFactor = %.2f, want >= %.2f", result.EaseFactor, MinEaseFactor)
	}
	if !approxEqual(result.EaseFactor, MinEaseFactor) {
		t.Errorf("easeFactor = %.2f, want %.2f (clamped)", result.EaseFactor, MinEaseFactor)
	}
}

func TestProcessReview_ReviewRepsIncrement(t *testing.T) {
	state := newState("review", 2.5, 10, 0, 5)

	goodResult := ProcessReview(state, Good)
	if goodResult.Repetitions != 6 {
		t.Errorf("Good: repetitions = %d, want 6", goodResult.Repetitions)
	}

	easyResult := ProcessReview(state, Easy)
	if easyResult.Repetitions != 6 {
		t.Errorf("Easy: repetitions = %d, want 6", easyResult.Repetitions)
	}

	// Hard and Again don't increment reps.
	hardResult := ProcessReview(state, Hard)
	if hardResult.Repetitions != 5 {
		t.Errorf("Hard: repetitions = %d, want 5", hardResult.Repetitions)
	}

	againResult := ProcessReview(state, Again)
	if againResult.Repetitions != 5 {
		t.Errorf("Again: repetitions = %d, want 5", againResult.Repetitions)
	}
}

func TestProcessReview_InputNotMutated(t *testing.T) {
	state := newState("review", 2.5, 10, 0, 3)
	origStatus := state.Status
	origEase := state.EaseFactor
	origInterval := state.IntervalDays

	_ = ProcessReview(state, Good)

	if state.Status != origStatus {
		t.Error("Status was mutated")
	}
	if state.EaseFactor != origEase {
		t.Error("EaseFactor was mutated")
	}
	if state.IntervalDays != origInterval {
		t.Error("IntervalDays was mutated")
	}
}

func TestClampInterval(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"below minimum clamps to 1", 0.5, 1.0},
		{"above maximum clamps to 365", 400.0, 365.0},
		{"normal value passes through", 10.0, 10.0},
		{"rounds to 1 decimal", 10.456, 10.5},
		{"rounds to 1 decimal down", 10.44, 10.4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampInterval(tt.in)
			if !approxEqual(got, tt.want) {
				t.Errorf("clampInterval(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatDays(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0.5, "1d"},
		{1, "1d"},
		{3, "3d"},
		{7, "1w"},
		{14, "2w"},
		{30, "1mo"},
		{90, "3mo"},
		{365, "1.0y"},
		{730, "2.0y"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDays(tt.in)
			if got != tt.want {
				t.Errorf("formatDays(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatMinutes(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{1, "1m"},
		{10, "10m"},
		{30, "30m"},
		{60, "1h"},
		{120, "2h"},
		{1440, "1d"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatMinutes(tt.in)
			if got != tt.want {
				t.Errorf("formatMinutes(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatInterval(t *testing.T) {
	tests := []struct {
		name   string
		state  *models.SRSState
		rating int
		want   string
	}{
		{
			name:   "new card Again shows 1m",
			state:  newState("new", 2.5, 0, 0, 0),
			rating: Again,
			want:   "1m",
		},
		{
			name:   "new card Good shows 10m",
			state:  newState("new", 2.5, 0, 0, 0),
			rating: Good,
			want:   "10m",
		},
		{
			name:   "new card Easy shows 4d",
			state:  newState("new", 2.5, 0, 0, 0),
			rating: Easy,
			want:   "4d",
		},
		{
			name:   "review card Good shows interval",
			state:  newState("review", 2.5, 10, 0, 3),
			rating: Good,
			want:   "4w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatInterval(tt.state, tt.rating)
			if got != tt.want {
				t.Errorf("FormatInterval() = %q, want %q", got, tt.want)
			}
		})
	}
}

func approxEqual(a, b float64) bool {
	const epsilon = 0.01
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}
