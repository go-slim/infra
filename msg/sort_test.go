package msg

import (
	"sort"
	"testing"
)

// LocaleByGenerality implements sort.Interface for []Locale based on generality hierarchy
// More general locales come first (they can contain more specific ones)
type LocaleByGenerality []Locale

func (lg LocaleByGenerality) Len() int           { return len(lg) }
func (lg LocaleByGenerality) Swap(i, j int)      { lg[i], lg[j] = lg[j], lg[i] }
func (lg LocaleByGenerality) Less(i, j int) bool { return lg.compare(lg[i], lg[j]) }

// compare returns true if locale i should come before locale j in sorting
func (lg LocaleByGenerality) compare(i, j Locale) bool {
	// If i contains j, then i is more general than j
	if i.Contains(j) && !j.Contains(i) {
		return true
	}
	// If j contains i, then j is more general than i
	if j.Contains(i) && !i.Contains(j) {
		return false
	}

	// Calculate generality levels: higher number = more specific
	iLevel := lg.generalityLevel(i)
	jLevel := lg.generalityLevel(j)

	// More general (lower level) comes first
	if iLevel != jLevel {
		return iLevel < jLevel
	}

	// Same generality level: use more sophisticated comparison
	return lg.compareSameLevel(i, j)
}

// generalityLevel returns a numeric level: 0=most general, 3=most specific
func (lg LocaleByGenerality) generalityLevel(l Locale) int {
	level := 0
	if l.Script() != "" {
		level++
	}
	if l.Region() != "" {
		level++
	}
	return level
}

// compareSameLevel compares locales at the same generality level
func (lg LocaleByGenerality) compareSameLevel(i, j Locale) bool {
	iScript, iRegion := i.Script(), i.Region()
	jScript, jRegion := j.Script(), j.Region()

	// If both have different types of components
	if iScript != "" && jRegion != "" && iRegion == "" && jScript == "" {
		// Script-based comes before Region-based at same level
		return true
	}
	if iRegion != "" && jScript != "" && iScript == "" && jRegion == "" {
		// Region-based comes after Script-based at same level
		return false
	}

	// Otherwise, use alphabetical order
	return string(i) < string(j)
}

func TestLocaleSort(t *testing.T) {
	locales := []Locale{
		"zh", "zh-CN", "zh-TW", "zh-Hans", "zh-Hant",
		"zh-Hans-CN", "zh-Hant-TW", "en",
	}

	// Expected order: most general to most specific
	// 1. Language only: en, zh (en < zh alphabetically)
	// 2. Language+Script: zh-Hans, zh-Hant (Hans < Hant)
	// 3. Language+Region: zh-CN, zh-TW (CN < TW)
	// 4. Language+Script+Region: zh-Hans-CN, zh-Hant-TW (Hans < Hant)
	expected := []Locale{
		"en", "zh", // Most general
		"zh-Hans", "zh-Hant", // Language+Script
		"zh-CN", "zh-TW", // Language+Region
		"zh-Hans-CN", "zh-Hant-TW", // Most specific
	}

	// Test our sorting
	sortable := LocaleByGenerality(locales)
	sort.Sort(sortable)

	t.Logf("Original order: %v", locales)
	t.Logf("Sorted order:   %v", []Locale(sortable))
	t.Logf("Expected order: %v", expected)

	// Verify the sorting result
	if !equalLocaleSlices([]Locale(sortable), expected) {
		t.Errorf("Sorting failed")
		t.Logf("Expected: %v", expected)
		t.Logf("Got:      %v", []Locale(sortable))
	}

	// Verify that Contains hierarchy is respected
	for i := 0; i < len(sortable)-1; i++ {
		for j := i + 1; j < len(sortable); j++ {
			if sortable[j].Contains(sortable[i]) && !sortable[i].Contains(sortable[j]) {
				t.Errorf("Hierarchy violation: more specific locale %v should not come before more general %v",
					sortable[j], sortable[i])
			}
		}
	}
}

func equalLocaleSlices(a, b []Locale) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Test specific sorting scenarios
func TestLocaleSort_Scenarios(t *testing.T) {
	tests := []struct {
		name     string
		input    []Locale
		expected []Locale
	}{
		{
			name:     "Chinese hierarchy",
			input:    []Locale{"zh-Hans-CN", "zh", "zh-Hans", "zh-CN"},
			expected: []Locale{"zh", "zh-Hans", "zh-CN", "zh-Hans-CN"},
		},
		{
			name:     "Mixed languages",
			input:    []Locale{"en-US", "zh-CN", "en", "zh"},
			expected: []Locale{"en", "zh", "en-US", "zh-CN"},
		},
		{
			name:     "Complex Chinese variants",
			input:    []Locale{"zh-Hant-TW", "zh-CN", "zh-Hans-CN", "zh", "zh-Hant", "zh-Hans", "zh-TW"},
			expected: []Locale{"zh", "zh-Hans", "zh-Hant", "zh-CN", "zh-TW", "zh-Hans-CN", "zh-Hant-TW"},
		},
		{
			name:     "English variants",
			input:    []Locale{"en-GB", "en-US", "en"},
			expected: []Locale{"en", "en-GB", "en-US"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortable := LocaleByGenerality(tt.input)
			sort.Sort(sortable)

			t.Logf("Input:    %v", tt.input)
			t.Logf("Sorted:   %v", []Locale(sortable))
			t.Logf("Expected: %v", tt.expected)

			if !equalLocaleSlices([]Locale(sortable), tt.expected) {
				t.Errorf("Sorting failed for %s", tt.name)
			}
		})
	}
}
