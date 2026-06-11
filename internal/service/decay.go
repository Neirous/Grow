package service

import (
	"math"
	"time"

	"grow/internal/models"
)

// CalculateEffectiveValue computes the decay-adjusted value.
// Returns the effective value clamped to never go below baseValue.
func CalculateEffectiveValue(currentValue, decayRate float64, lastActivityAt *time.Time, now time.Time, baseValue float64) float64 {
	if lastActivityAt == nil || decayRate <= 0 {
		return currentValue
	}
	days := now.Sub(*lastActivityAt).Hours() / 24
	if days <= 0 {
		return currentValue
	}
	effective := currentValue * math.Pow(1-decayRate/100, days)
	if effective < baseValue {
		effective = baseValue
	}
	return effective
}

// ApplyDecay populates the EffectiveValue and DaysSinceLastActivity fields on an Ability.
func ApplyDecay(a *models.Ability) {
	now := time.Now()
	if a.LastActivityAt != nil {
		a.DaysSinceLastActivity = now.Sub(*a.LastActivityAt).Hours() / 24
	}
	a.EffectiveValue = CalculateEffectiveValue(
		a.CurrentValue, a.DecayRate, a.LastActivityAt, now, a.BaseValue,
	)
}

// ApplyDecayToMany applies decay to a slice of abilities.
func ApplyDecayToMany(abilities []models.Ability) {
	for i := range abilities {
		ApplyDecay(&abilities[i])
	}
}
