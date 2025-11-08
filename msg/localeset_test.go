package msg

import (
	"testing"
)

func TestLocaleSet_IsUnlimited(t *testing.T) {
	tests := []struct {
		name     string
		locales  LocaleSet
		expected bool
	}{
		{
			name:     "Nil LocaleSet is unlimited",
			locales:  nil,
			expected: true,
		},
		{
			name:     "Empty slice is not unlimited",
			locales:  LocaleSet{},
			expected: false,
		},
		{
			name:     "Non-empty slice is not unlimited",
			locales:  LocaleSet{English, Chinese},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.locales.IsUnlimited() != tt.expected {
				t.Errorf("LocaleSet.IsUnlimited() = %v, want %v", tt.locales.IsUnlimited(), tt.expected)
			}
		})
	}
}

func TestLocaleSet_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		locales  LocaleSet
		expected bool
	}{
		{
			name:     "Nil LocaleSet is not empty",
			locales:  nil,
			expected: false,
		},
		{
			name:     "Empty slice is empty",
			locales:  LocaleSet{},
			expected: true,
		},
		{
			name:     "Non-empty slice is not empty",
			locales:  LocaleSet{English, Chinese},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.locales.IsEmpty() != tt.expected {
				t.Errorf("LocaleSet.IsEmpty() = %v, want %v", tt.locales.IsEmpty(), tt.expected)
			}
		})
	}
}

func TestLocaleSet_ExactContains(t *testing.T) {
	tests := []struct {
		name     string
		locales  LocaleSet
		test     Locale
		expected bool
	}{
		{
			name:     "Unlimited contains all",
			locales:  nil,
			test:     Locale("any-locale"),
			expected: true,
		},
		{
			name:     "Empty contains none",
			locales:  LocaleSet{},
			test:     English,
			expected: false,
		},
		{
			name:     "Exact match",
			locales:  LocaleSet{English, Chinese},
			test:     English,
			expected: true,
		},
		{
			name:     "Parent contains child",
			locales:  LocaleSet{Chinese},
			test:     Locale("zh-CN"),
			expected: true,
		},
		{
			name:     "Child contains parent",
			locales:  LocaleSet{Locale("zh-CN")},
			test:     Chinese,
			expected: true,
		},
		{
			name:     "Script match",
			locales:  LocaleSet{ChineseSimplified},
			test:     Locale("zh-Hans-CN"),
			expected: true,
		},
		{
			name:     "No match",
			locales:  LocaleSet{English, French},
			test:     Chinese,
			expected: false,
		},
		{
			name:     "Complex match",
			locales:  LocaleSet{Locale("zh-Hans")},
			test:     Locale("zh-Hans-CN"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.locales.Contains(tt.test) != tt.expected {
				t.Errorf("LocaleSet.Contains(%q) = %v, want %v",
					string(tt.test), tt.locales.Contains(tt.test), tt.expected)
			}
		})
	}
}

func TestLocaleSet_Contains(t *testing.T) {
	tests := []struct {
		name     string
		locales  LocaleSet
		test     Locale
		expected bool
	}{
		{
			name:     "Unlimited contains all",
			locales:  nil,
			test:     Locale("any-locale"),
			expected: true,
		},
		{
			name:     "Empty contains none",
			locales:  LocaleSet{},
			test:     English,
			expected: false,
		},
		{
			name:     "Exact match",
			locales:  LocaleSet{English, Chinese},
			test:     English,
			expected: true,
		},
		{
			name:     "Parent contains child (loose matching)",
			locales:  LocaleSet{Chinese},
			test:     Locale("zh-CN"),
			expected: true,
		},
		{
			name:     "Child matches parent (loose matching)",
			locales:  LocaleSet{Locale("zh-CN")},
			test:     Chinese,
			expected: true,
		},
		{
			name:     "No match",
			locales:  LocaleSet{English, French},
			test:     Chinese,
			expected: false,
		},
		{
			name:     "Different script",
			locales:  LocaleSet{ChineseSimplified},
			test:     ChineseTraditional,
			expected: false,
		},
		{
			name:     "Same locale different instances",
			locales:  LocaleSet{Locale("en-US")},
			test:     Locale("en-US"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.locales.Contains(tt.test) != tt.expected {
				t.Errorf("LocaleSet.Contains(%q) = %v, want %v",
					string(tt.test), tt.locales.Contains(tt.test), tt.expected)
			}
		})
	}
}

func TestLocaleSet_Slice(t *testing.T) {
	tests := []struct {
		name     string
		locales  LocaleSet
		expected []Locale
	}{
		{
			name:     "Unlimited slice",
			locales:  nil,
			expected: nil,
		},
		{
			name:     "Empty slice",
			locales:  LocaleSet{},
			expected: []Locale{},
		},
		{
			name:     "Single locale",
			locales:  LocaleSet{English},
			expected: []Locale{English},
		},
		{
			name:     "Multiple locales",
			locales:  LocaleSet{English, Chinese, French},
			expected: []Locale{English, Chinese, French},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.locales.Slice()

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("LocaleSet.Slice() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			// Check contents
			for i, expected := range tt.expected {
				if i >= len(result) || !result[i].Equal(expected) {
					t.Errorf("LocaleSet.Slice()[%d] = %q, want %q", i, string(result[i]), string(expected))
				}
			}
		})
	}
}

func TestLocaleSet_ModificationSafety(t *testing.T) {
	t.Run("Slice modification safety", func(t *testing.T) {
		original := LocaleSet{English, Chinese, French}
		slice := original.Slice()

		// Modify the slice
		slice[0] = Locale("modified")
		slice = append(slice, Spanish)

		// Original should be unchanged
		if !original[0].Equal(English) {
			t.Error("Modifying slice should not affect original LocaleSet")
		}

		if len(original) != 3 {
			t.Error("Appending to slice should not affect original LocaleSet length")
		}
	})

	t.Run("Multiple calls return independent slices", func(t *testing.T) {
		original := LocaleSet{English, Chinese}
		slice1 := original.Slice()
		slice2 := original.Slice()

		// Modify one slice
		slice1[0] = Locale("modified")

		// Other slice should be unchanged
		if !slice2[0].Equal(English) {
			t.Error("Multiple Slice() calls should return independent slices")
		}
	})
}

func TestLocaleSet_EdgeCases(t *testing.T) {
	t.Run("Empty locale in set", func(t *testing.T) {
		locales := LocaleSet{English, Locale(""), Chinese}

		if !locales.Contains(Locale("")) {
			t.Error("LocaleSet should contain empty locale")
		}

		if !locales.Contains(Locale("")) {
			t.Error("LocaleSet should loosely contain empty locale")
		}
	})

	t.Run("Invalid locale in set", func(t *testing.T) {
		invalid := Locale("invalid-xyz-123")
		locales := LocaleSet{English, invalid}

		if !locales.Contains(invalid) {
			t.Error("LocaleSet should contain invalid locale")
		}

		if !locales.Contains(invalid) {
			t.Error("LocaleSet should loosely contain invalid locale")
		}
	})

	t.Run("Unicode locale in set", func(t *testing.T) {
		unicode := Locale("en-US-🏴󠁧󠁢󠁳󠁣󠁴󠁿")
		locales := LocaleSet{unicode}

		if !locales.Contains(unicode) {
			t.Error("LocaleSet should contain Unicode locale")
		}

		if !locales.Contains(unicode) {
			t.Error("LocaleSet should loosely contain Unicode locale")
		}
	})
}

func TestLocaleSet_Performance(t *testing.T) {
	t.Run("Large set performance", func(t *testing.T) {
		// Create a large set
		locales := make(LocaleSet, 1000)
		for i := 0; i < 1000; i++ {
			locales[i] = Locale("locale-" + string(rune(i)))
		}

		testLocale := Locale("locale-500")

		// Test Contains performance
		for i := 0; i < 100; i++ {
			_ = locales.Contains(testLocale)
		}

		// Test Contains performance
		for i := 0; i < 100; i++ {
			_ = locales.Contains(testLocale)
		}
		// If we reach here without timeout, performance is acceptable
	})
}

func TestLocaleSet_RealWorldScenarios(t *testing.T) {
	t.Run("Typical language support", func(t *testing.T) {
		supported := LocaleSet{English, Chinese, Spanish, French, German, Japanese}

		// Test loose matches (Contains now does loose matching)
		testCases := []struct {
			locale   Locale
			expected bool
		}{
			{English, true},
			{Locale("en-US"), true}, // Child of English - should match
			{Locale("zh-CN"), true}, // Child of Chinese - should match
			{Russian, false},
			{Korean, false},
		}

		for _, tc := range testCases {
			if supported.Contains(tc.locale) != tc.expected {
				t.Errorf("Contains(%q) = %v, want %v",
					string(tc.locale), supported.Contains(tc.locale), tc.expected)
			}
		}
	})

	t.Run("Regional language support", func(t *testing.T) {
		supported := LocaleSet{Locale("en-US"), Locale("en-GB"), Locale("zh-CN"), Locale("zh-TW")}

		testCases := []struct {
			locale   Locale
			expected bool
		}{
			{Locale("en-US"), true},
			{Locale("en-GB"), true},
			{English, true}, // Parent of supported locales - should match with loose logic
			{Locale("zh-CN"), true},
			{Chinese, true}, // Parent of supported locales - should match with loose logic
			{Locale("fr-FR"), false},
		}

		for _, tc := range testCases {
			if supported.Contains(tc.locale) != tc.expected {
				t.Errorf("Contains(%q) = %v, want %v",
					string(tc.locale), supported.Contains(tc.locale), tc.expected)
			}
		}
	})
}

// Additional tests for LooseContains behavior from localeset_loose_test.go

func TestLocaleSet_Contains_ParentChildRelationships(t *testing.T) {
	tests := []struct {
		name     string
		set      LocaleSet
		test     Locale
		expected bool
		reason   string
	}{
		{
			name:     "Parent language contains child region",
			set:      LocaleSet{Chinese},
			test:     Locale("zh-CN"),
			expected: true,
			reason:   "zh contains zh-CN",
		},
		{
			name:     "Child region contains parent language",
			set:      LocaleSet{Locale("zh-CN")},
			test:     Chinese,
			expected: true,
			reason:   "zh-CN contains zh",
		},
		{
			name:     "Parent language contains child script",
			set:      LocaleSet{Chinese},
			test:     ChineseSimplified,
			expected: true,
			reason:   "zh contains zh-Hans-CN",
		},
		{
			name:     "Child script contains parent language",
			set:      LocaleSet{ChineseSimplified},
			test:     Chinese,
			expected: true,
			reason:   "zh-Hans-CN contains zh",
		},
		{
			name:     "Script contains script+region",
			set:      LocaleSet{ChineseSimplified},
			test:     Locale("zh-Hans-CN"),
			expected: true,
			reason:   "zh-Hans-CN contains zh-Hans-CN (same)",
		},
		{
			name:     "Script+region contains script",
			set:      LocaleSet{Locale("zh-Hans-CN")},
			test:     ChineseSimplified,
			expected: true,
			reason:   "zh-Hans-CN contains zh-Hans-CN (same)",
		},
		{
			name:     "English contains en-US",
			set:      LocaleSet{English},
			test:     Locale("en-US"),
			expected: true,
			reason:   "en contains en-US",
		},
		{
			name:     "en-US contains English",
			set:      LocaleSet{Locale("en-US")},
			test:     English,
			expected: true,
			reason:   "en-US contains en",
		},
		{
			name:     "English contains en-GB",
			set:      LocaleSet{English},
			test:     Locale("en-GB"),
			expected: true,
			reason:   "en contains en-GB",
		},
		{
			name:     "en-GB contains English",
			set:      LocaleSet{Locale("en-GB")},
			test:     English,
			expected: true,
			reason:   "en-GB contains en",
		},
		{
			name:     "en-US does not contain en-GB",
			set:      LocaleSet{Locale("en-US")},
			test:     Locale("en-GB"),
			expected: false,
			reason:   "en-US does not contain en-GB (different regions)",
		},
		{
			name:     "en-GB does not contain en-US",
			set:      LocaleSet{Locale("en-GB")},
			test:     Locale("en-US"),
			expected: false,
			reason:   "en-GB does not contain en-US (different regions)",
		},
		{
			name:     "zh-Hans does not contain zh-Hant",
			set:      LocaleSet{ChineseSimplified},
			test:     ChineseTraditional,
			expected: false,
			reason:   "zh-Hans-CN does not contain zh-Hant-TW (different scripts)",
		},
		{
			name:     "zh-Hant does not contain zh-Hans",
			set:      LocaleSet{ChineseTraditional},
			test:     ChineseSimplified,
			expected: false,
			reason:   "zh-Hant-TW does not contain zh-Hans-CN (different scripts)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.set.Contains(tt.test)
			if result != tt.expected {
				t.Errorf("LocaleSet.Contains(%q) = %v, want %v. Reason: %s",
					string(tt.test), result, tt.expected, tt.reason)
			}
		})
	}
}

func TestLocaleSet_Contains_ComplexHierarchies(t *testing.T) {
	tests := []struct {
		name     string
		set      LocaleSet
		test     Locale
		expected bool
	}{
		{
			name:     "zh contains zh-Hans-CN-TW (invalid but test parsing)",
			set:      LocaleSet{Chinese},
			test:     Locale("zh-Hans-CN-TW"),
			expected: true,
		},
		{
			name:     "zh-Hans contains zh-Hans-CN",
			set:      LocaleSet{Locale("zh-Hans")},
			test:     Locale("zh-Hans-CN"),
			expected: true,
		},
		{
			name:     "zh-Hans-CN contains zh-Hans",
			set:      LocaleSet{Locale("zh-Hans-CN")},
			test:     Locale("zh-Hans"),
			expected: true,
		},
		{
			name:     "sr-Latn contains sr-Latn-RS",
			set:      LocaleSet{Locale("sr-Latn")},
			test:     Locale("sr-Latn-RS"),
			expected: true,
		},
		{
			name:     "sr-Latn-RS contains sr-Latn",
			set:      LocaleSet{Locale("sr-Latn-RS")},
			test:     Locale("sr-Latn"),
			expected: true,
		},
		{
			name:     "sr contains sr-Latn",
			set:      LocaleSet{Locale("sr")},
			test:     Locale("sr-Latn"),
			expected: true,
		},
		{
			name:     "sr-Latn contains sr",
			set:      LocaleSet{Locale("sr-Latn")},
			test:     Locale("sr"),
			expected: true,
		},
		{
			name:     "sr-Latn does not contain sr-Cyrl",
			set:      LocaleSet{Locale("sr-Latn")},
			test:     Locale("sr-Cyrl"),
			expected: false,
		},
		{
			name:     "sr-Cyrl does not contain sr-Latn",
			set:      LocaleSet{Locale("sr-Cyrl")},
			test:     Locale("sr-Latn"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.set.Contains(tt.test)
			if result != tt.expected {
				t.Errorf("LocaleSet.Contains(%q) = %v, want %v",
					string(tt.test), result, tt.expected)
			}
		})
	}
}

func TestLocaleSet_Contains_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		set      LocaleSet
		test     Locale
		expected bool
	}{
		{
			name:     "Empty locale contains empty locale",
			set:      LocaleSet{Locale("")},
			test:     Locale(""),
			expected: true,
		},
		{
			name:     "Empty locale contains non-empty locale",
			set:      LocaleSet{Locale("")},
			test:     English,
			expected: false,
		},
		{
			name:     "Non-empty locale contains empty locale",
			set:      LocaleSet{English},
			test:     Locale(""),
			expected: false,
		},
		{
			name:     "Invalid locale contains same invalid locale",
			set:      LocaleSet{Locale("invalid-xyz")},
			test:     Locale("invalid-xyz"),
			expected: true,
		},
		{
			name:     "Invalid locale does not contain other invalid locale",
			set:      LocaleSet{Locale("invalid-xyz")},
			test:     Locale("different-invalid"),
			expected: false,
		},
		{
			name:     "Private use locale contains same private use",
			set:      LocaleSet{Locale("x-private")},
			test:     Locale("x-private"),
			expected: true,
		},
		{
			name:     "x-unknown contains standard locale",
			set:      LocaleSet{Locale("x-unknown")},
			test:     English,
			expected: false,
		},
		{
			name:     "Standard locale does not contain x-unknown",
			set:      LocaleSet{English},
			test:     Locale("x-unknown"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.set.Contains(tt.test)
			if result != tt.expected {
				t.Errorf("LocaleSet.Contains(%q) = %v, want %v",
					string(tt.test), result, tt.expected)
			}
		})
	}
}

func TestLocaleSet_Contains_RealWorldLanguageFamilies(t *testing.T) {
	// Test real-world language family relationships
	tests := []struct {
		name     string
		set      LocaleSet
		test     Locale
		expected bool
		scenario string
	}{
		// Chinese family
		{
			name:     "Chinese family - zh contains zh-CN",
			set:      LocaleSet{Chinese},
			test:     Locale("zh-CN"),
			expected: true,
			scenario: "Chinese language contains Simplified Chinese region",
		},
		{
			name:     "Chinese family - zh-CN contains zh",
			set:      LocaleSet{Locale("zh-CN")},
			test:     Chinese,
			expected: true,
			scenario: "Simplified Chinese region contains Chinese language",
		},
		{
			name:     "Chinese family - zh-Hans contains zh-Hans-CN",
			set:      LocaleSet{ChineseSimplified},
			test:     Locale("zh-Hans-CN"),
			expected: true,
			scenario: "Simplified Chinese script contains Simplified Chinese region",
		},
		{
			name:     "Chinese family - zh-Hans-CN contains zh-Hans",
			set:      LocaleSet{Locale("zh-Hans-CN")},
			test:     Locale("zh-Hans"),
			expected: true,
			scenario: "Simplified Chinese region contains Simplified Chinese script",
		},
		// English family
		{
			name:     "English family - en contains en-US",
			set:      LocaleSet{English},
			test:     Locale("en-US"),
			expected: true,
			scenario: "English language contains US English region",
		},
		{
			name:     "English family - en-US contains en-GB",
			set:      LocaleSet{Locale("en-US")},
			test:     Locale("en-GB"),
			expected: false,
			scenario: "US English does not contain British English (different regions)",
		},
		// Spanish family
		{
			name:     "Spanish family - es contains es-ES",
			set:      LocaleSet{Spanish},
			test:     Locale("es-ES"),
			expected: true,
			scenario: "Spanish language contains Spain Spanish region",
		},
		{
			name:     "Spanish family - es contains es-MX",
			set:      LocaleSet{Spanish},
			test:     Locale("es-MX"),
			expected: true,
			scenario: "Spanish language contains Mexican Spanish region",
		},
		{
			name:     "Spanish family - es-ES does not contain es-MX",
			set:      LocaleSet{Locale("es-ES")},
			test:     Locale("es-MX"),
			expected: false,
			scenario: "Spain Spanish does not contain Mexican Spanish (different regions)",
		},
		// Cross-family tests
		{
			name:     "Cross family - en does not contain zh",
			set:      LocaleSet{English},
			test:     Chinese,
			expected: false,
			scenario: "English does not contain Chinese (different languages)",
		},
		{
			name:     "Cross family - zh does not contain en",
			set:      LocaleSet{Chinese},
			test:     English,
			expected: false,
			scenario: "Chinese does not contain English (different languages)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.set.Contains(tt.test)
			if result != tt.expected {
				t.Errorf("LocaleSet.Contains(%q) = %v, want %v. Scenario: %s",
					string(tt.test), result, tt.expected, tt.scenario)
			}
		})
	}
}

func TestLocaleSet_Contains_PerformanceAndComplexity(t *testing.T) {
	// Test with large sets to ensure performance is acceptable
	t.Run("Large set performance", func(t *testing.T) {
		// Create a set with many locales
		locales := make(LocaleSet, 100)
		baseLanguages := []Locale{English, Chinese, Spanish, French, German, Japanese, Korean, Russian}

		for i := 0; i < 100; i++ {
			lang := baseLanguages[i%len(baseLanguages)]
			if i < 20 {
				locales[i] = lang
			} else if i < 40 {
				locales[i] = Locale(string(lang) + "-US")
			} else if i < 60 {
				locales[i] = Locale(string(lang) + "-CN")
			} else if i < 80 {
				locales[i] = Locale(string(lang) + "-Hans")
			} else {
				locales[i] = Locale(string(lang) + "-Hans-CN")
			}
		}

		// Test various lookups
		testCases := []Locale{
			English,
			Locale("en-US"),
			Chinese,
			Locale("zh-CN"),
			Locale("zh-Hans"),
			Locale("zh-Hans-CN"),
			Spanish,
			Locale("es-MX"),
			French,
			Locale("fr-FR"),
		}

		for _, test := range testCases {
			// This should complete quickly even with a large set
			_ = locales.Contains(test)
		}
		// If we reach here without timeout, performance is acceptable
	})
}

func TestLocaleSet_Contains_UnicodeAndSpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		set      LocaleSet
		test     Locale
		expected bool
	}{
		{
			name:     "Unicode locale contains same Unicode locale",
			set:      LocaleSet{Locale("en-US-🏴󠁧󠁢󠁳󠁣󠁴󠁿")},
			test:     Locale("en-US-🏴󠁧󠁢󠁳󠁣󠁴󠁿"),
			expected: true,
		},
		{
			name:     "Base does not contain Unicode variant",
			set:      LocaleSet{Locale("en-US")},
			test:     Locale("en-US-🏴󠁧󠁢󠁳󠁣󠁴󠁿"),
			expected: false,
		},
		{
			name:     "Unicode variant does not contain base",
			set:      LocaleSet{Locale("en-US-🏴󠁧󠁢󠁳󠁣󠁴󠁿")},
			test:     Locale("en-US"),
			expected: false,
		},
		{
			name:     "Private use with Unicode contains base",
			set:      LocaleSet{Locale("en-x-emoji-🎉")},
			test:     English,
			expected: true,
		},
		{
			name:     "Base contains private use with Unicode",
			set:      LocaleSet{English},
			test:     Locale("en-x-emoji-🎉"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.set.Contains(tt.test)
			if result != tt.expected {
				t.Errorf("LocaleSet.Contains(%q) = %v, want %v",
					string(tt.test), result, tt.expected)
			}
		})
	}
}

func TestLocaleSet_Contains_BoundaryConditions(t *testing.T) {
	// Test boundary conditions and edge cases
	t.Run("Maximum nesting levels", func(t *testing.T) {
		// Test with very deeply nested locales
		deepLocale := Locale("a-b-c-d-e-f-g-h-i-j-k-l-m-n-o-p-q-r-s-t-u-v-w-x-y-z")
		shallowLocale := Locale("a")

		tests := []struct {
			set      LocaleSet
			test     Locale
			expected bool
		}{
			{LocaleSet{shallowLocale}, deepLocale, true},
			{LocaleSet{deepLocale}, shallowLocale, true},
			{LocaleSet{deepLocale}, deepLocale, true},
			{LocaleSet{shallowLocale}, shallowLocale, true},
		}

		for _, tt := range tests {
			result := tt.set.Contains(tt.test)
			if result != tt.expected {
				t.Errorf("LocaleSet.Contains(%q) = %v, want %v",
					string(tt.test), result, tt.expected)
			}
		}
	})

	t.Run("Single character components", func(t *testing.T) {
		tests := []struct {
			set      LocaleSet
			test     Locale
			expected bool
		}{
			{LocaleSet{Locale("a")}, Locale("a-b"), true},
			{LocaleSet{Locale("a-b")}, Locale("a"), true},
			{LocaleSet{Locale("a-b-c")}, Locale("a"), true},
			{LocaleSet{Locale("a")}, Locale("a-b-c"), true},
			{LocaleSet{Locale("a-b")}, Locale("a-b-c"), false},
			{LocaleSet{Locale("a-b-c")}, Locale("a-b"), false},
		}

		for _, tt := range tests {
			result := tt.set.Contains(tt.test)
			if result != tt.expected {
				t.Errorf("LocaleSet.Contains(%q) = %v, want %v",
					string(tt.test), result, tt.expected)
			}
		}
	})
}
