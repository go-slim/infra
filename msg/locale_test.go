package msg

import (
	"fmt"
	"testing"
)

func TestLocale_NewLocale(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"en", "en"},
		{"zh-CN", "zh-CN"},
		{"zh-Hans-CN", "zh-Hans-CN"},
		{"en-US", "en-US"},
		{"fr-FR", "fr-FR"},
		{"ja", "ja"},
		{"ko-KR", "ko-KR"},
		{"pt-BR", "pt-BR"},
		{"ru-RU", "ru-RU"},
		{"ar-SA", "ar-SA"},
		{"hi-IN", "hi-IN"},
		{"th-TH", "th-TH"},
		{"vi-VN", "vi-VN"},
		{"", ""},
		{"invalid-locale", "invalid-locale"},
		{"x-unknown", "x-unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			locale := Locale(tt.input)
			if string(locale) != tt.expected {
				t.Errorf("Locale(%q) = %q, want %q", tt.input, string(locale), tt.expected)
			}
		})
	}
}

func TestLocale_Language(t *testing.T) {
	tests := []struct {
		name     string
		locale   Locale
		expected string
	}{
		{"Simple language", English, "en"},
		{"Chinese Simplified", ChineseSimplified, "zh"},
		{"English US", Locale("en-US"), "en"},
		{"Japanese", Japanese, "ja"},
		{"Spanish", Spanish, "es"},
		{"French", French, "fr"},
		{"German", German, "de"},
		{"Chinese Traditional", ChineseTraditional, "zh"},
		{"Empty locale", Locale(""), ""},
		{"Invalid locale", Locale("invalid-xyz"), "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.locale.Language() != tt.expected {
				t.Errorf("Locale(%q).Language() = %q, want %q", string(tt.locale), tt.locale.Language(), tt.expected)
			}
		})
	}
}

func TestLocale_Script(t *testing.T) {
	tests := []struct {
		name     string
		locale   Locale
		expected string
	}{
		{"Simple language (no script)", English, ""},
		{"Chinese Simplified", ChineseSimplified, "Hans"},
		{"Chinese Traditional", ChineseTraditional, "Hant"},
		{"Japanese with script", Locale("ja-Jpan"), "Jpan"},
		{"Serbian Cyrillic", Locale("sr-Cyrl"), "Cyrl"},
		{"Serbian Latin", Locale("sr-Latn"), "Latn"},
		{"Mongolian Cyrillic", Locale("mn-Cyrl"), "Cyrl"},
		{"Mongolian Mongolian", Locale("mn-Mong"), "Mong"},
		{"Empty locale", Locale(""), ""},
		{"No script specified", Locale("en-US"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.locale.Script() != tt.expected {
				t.Errorf("Locale(%q).Script() = %q, want %q", string(tt.locale), tt.locale.Script(), tt.expected)
			}
		})
	}
}

func TestLocale_Region(t *testing.T) {
	tests := []struct {
		name     string
		locale   Locale
		expected string
	}{
		{"Simple language (no region)", English, ""},
		{"English US", Locale("en-US"), "US"},
		{"Chinese Simplified", ChineseSimplified, ""},
		{"Chinese Traditional", ChineseTraditional, ""},
		{"Japanese Japan", Locale("ja-JP"), "JP"},
		{"Spanish Spain", Locale("es-ES"), "ES"},
		{"Spanish Mexico", Locale("es-MX"), "MX"},
		{"French France", Locale("fr-FR"), "FR"},
		{"French Canada", Locale("fr-CA"), "CA"},
		{"German Germany", Locale("de-DE"), "DE"},
		{"Portuguese Brazil", Locale("pt-BR"), "BR"},
		{"Russian Russia", Locale("ru-RU"), "RU"},
		{"Empty locale", Locale(""), ""},
		{"No region specified", Locale("zh"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.locale.Region() != tt.expected {
				t.Errorf("Locale(%q).Region() = %q, want %q", string(tt.locale), tt.locale.Region(), tt.expected)
			}
		})
	}
}

func TestLocale_String(t *testing.T) {
	tests := []struct {
		name     string
		locale   Locale
		expected string
	}{
		{"English", English, "en"},
		{"Chinese Simplified", ChineseSimplified, "zh-Hans"},
		{"English US", Locale("en-US"), "en-US"},
		{"Full locale", Locale("zh-Hans-CN"), "zh-Hans-CN"},
		{"Empty locale", Locale(""), ""},
		{"Invalid locale", Locale("invalid-xyz"), "invalid-xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.locale.String() != tt.expected {
				t.Errorf("Locale(%q).String() = %q, want %q", string(tt.locale), tt.locale.String(), tt.expected)
			}
		})
	}
}

func TestLocale_Equals(t *testing.T) {
	tests := []struct {
		name     string
		locale1  Locale
		locale2  Locale
		expected bool
	}{
		{"Same locale", English, English, true},
		{"Same string", Locale("en-US"), Locale("en-US"), true},
		{"Different locales", English, Chinese, false},
		{"Empty vs non-empty", Locale(""), English, false},
		{"Both empty", Locale(""), Locale(""), true},
		{"Case sensitive", Locale("en-US"), Locale("en-us"), false},
		{"Chinese variants", ChineseSimplified, ChineseTraditional, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.locale1.Equal(tt.locale2) != tt.expected {
				t.Errorf("Locale(%q).Equal(%q) = %v, want %v",
					string(tt.locale1), string(tt.locale2), tt.locale1.Equal(tt.locale2), tt.expected)
			}
		})
	}
}

func TestLocale_Contains(t *testing.T) {
	tests := []struct {
		name     string
		parent   Locale
		child    Locale
		expected bool
	}{
		// 更通用的包含更具体的
		{"Chinese contains zh-CN", Chinese, Locale("zh-CN"), true},
		{"Chinese contains zh-TW", Chinese, Locale("zh-TW"), true},
		{"Chinese contains zh-Hans", Chinese, ChineseSimplified, true},
		{"English contains en-US", English, Locale("en-US"), true},
		{"English contains en-GB", English, Locale("en-GB"), true},
		{"zh-Hans contains zh-Hans-CN", ChineseSimplified, Locale("zh-Hans-CN"), true},
		{"zh-CN contains zh-Hans-CN", Locale("zh-CN"), Locale("zh-Hans-CN"), true},
		{"zh-Hans contains zh-Hans-SG", ChineseSimplified, Locale("zh-Hans-SG"), true},
		{"zh-CN contains zh-Latn-CN", Locale("zh-CN"), Locale("zh-Latn-CN"), true},

		// 更具体的不能包含更通用的
		{"zh-Hans-CN contains zh", Locale("zh-Hans-CN"), Chinese, false},
		{"zh-Hans-CN contains zh-Hans", Locale("zh-Hans-CN"), ChineseSimplified, false},
		{"zh-Hans-CN contains zh-CN", Locale("zh-Hans-CN"), Locale("zh-CN"), false},
		{"en-US contains en", Locale("en-US"), English, false},

		// 相同具体程度但组件不同的
		{"zh-CN contains zh-TW", Locale("zh-CN"), Locale("zh-TW"), false},
		{"zh-CN contains zh-SG", Locale("zh-CN"), Locale("zh-SG"), false},
		{"zh-TW contains zh-CN", Locale("zh-TW"), Locale("zh-CN"), false},
		{"zh-Hans contains zh-Hant", ChineseSimplified, ChineseTraditional, false},
		{"zh-Hant contains zh-Hans", ChineseTraditional, ChineseSimplified, false},

		// 语言不同
		{"zh-CN contains en-US", Locale("zh-CN"), Locale("en-US"), false},
		{"en-US contains zh-CN", Locale("en-US"), Locale("zh-CN"), false},

		// 边界情况
		{"Exact match", English, English, true},
		{"Empty contains non-empty", Locale(""), English, false},
		{"Non-empty contains empty", English, Locale(""), false},
		{"Both empty", Locale(""), Locale(""), true},

		// 脚本匹配验证
		{"zh-Hans contains zh-Hans-CN", ChineseSimplified, Locale("zh-Hans-CN"), true},
		{"zh-Hans contains zh-Hant", ChineseSimplified, ChineseTraditional, false},
		{"zh-Hans contains zh-Hans-TW", ChineseSimplified, Locale("zh-Hans-TW"), true},

		// 地区匹配验证
		{"zh-CN contains zh-Hans-CN", Locale("zh-CN"), Locale("zh-Hans-CN"), true},
		{"zh-CN contains zh-CN", Locale("zh-CN"), Locale("zh-CN"), true},
		{"zh-CN contains zh-Latn-CN", Locale("zh-CN"), Locale("zh-Latn-CN"), true},
		{"zh-CN contains zh-TW", Locale("zh-CN"), Locale("zh-TW"), false},

		// 多语言测试
		{"en contains en-US", English, Locale("en-US"), true},
		{"en contains en-GB", English, Locale("en-GB"), true},
		{"en-US contains en-GB", Locale("en-US"), Locale("en-GB"), false},
		{"en-US contains en", Locale("en-US"), English, false},
		{"fr contains fr-FR", Locale("fr"), Locale("fr-FR"), true},
		{"fr-FR contains fr-Latn-FR", Locale("fr-FR"), Locale("fr-Latn-FR"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.parent.Contains(tt.child) != tt.expected {
				t.Errorf("Locale(%q).Contains(%q) = %v, want %v",
					string(tt.parent), string(tt.child), tt.parent.Contains(tt.child), tt.expected)
			}
		})
	}
}

func TestLocale_Constants(t *testing.T) {
	tests := []struct {
		name     string
		locale   Locale
		expected string
	}{
		{"English constant", English, "en"},
		{"Chinese constant", Chinese, "zh"},
		{"Chinese Simplified", ChineseSimplified, "zh-Hans"},
		{"Chinese Traditional", ChineseTraditional, "zh-Hant"},
		{"Spanish", Spanish, "es"},
		{"French", French, "fr"},
		{"German", German, "de"},
		{"Japanese", Japanese, "ja"},
		{"Korean", Korean, "ko"},
		{"Portuguese", Portuguese, "pt"},
		{"Russian", Russian, "ru"},
		{"Arabic", Arabic, "ar"},
		{"Hindi", Hindi, "hi"},
		{"Thai", Thai, "th"},
		{"Vietnamese", Vietnamese, "vi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.locale) != tt.expected {
				t.Errorf("Constant %s = %q, want %q", tt.name, string(tt.locale), tt.expected)
			}
		})
	}
}

func TestLocale_EdgeCases(t *testing.T) {
	t.Run("Empty locale methods", func(t *testing.T) {
		empty := Locale("")

		if empty.Language() != "" {
			t.Errorf("Empty locale language = %q, want empty string", empty.Language())
		}

		if empty.Script() != "" {
			t.Errorf("Empty locale script = %q, want empty string", empty.Script())
		}

		if empty.Region() != "" {
			t.Errorf("Empty locale region = %q, want empty string", empty.Region())
		}

		if !empty.Equal(Locale("")) {
			t.Error("Empty locale should equal empty locale")
		}
	})

	t.Run("Invalid locale parsing", func(t *testing.T) {
		invalid := Locale("invalid-xyz-123")

		// Invalid locales should still be stored as-is
		if string(invalid) != "invalid-xyz-123" {
			t.Errorf("Invalid locale string = %q, want original", string(invalid))
		}

		// Language() should return the language code part for invalid locales
		if invalid.Language() != "invalid" {
			t.Errorf("Invalid locale language = %q, want language code only", invalid.Language())
		}
	})

	t.Run("Complex locale components", func(t *testing.T) {
		complex := Locale("sr-Latn-RS") // Serbian in Latin script in Serbia

		if complex.Language() != "sr" {
			t.Errorf("Complex locale language = %q, want sr", complex.Language())
		}

		if complex.Script() != "Latn" {
			t.Errorf("Complex locale script = %q, want Latn", complex.Script())
		}

		if complex.Region() != "RS" {
			t.Errorf("Complex locale region = %q, want RS", complex.Region())
		}
	})
}

func TestLocale_UnicodeAndSpecialCases(t *testing.T) {
	t.Run("Emoji and Unicode in locale", func(t *testing.T) {
		// Test that locales with Unicode characters are handled properly
		unicodeLocale := Locale("en-US-🏴󠁧󠁢󠁳󠁣󠁴󠁿") // en-US with Scotland flag

		if string(unicodeLocale) != "en-US-🏴󠁧󠁢󠁳󠁣󠁴󠁿" {
			t.Errorf("Unicode locale string not preserved")
		}
	})

	t.Run("Private use subtags", func(t *testing.T) {
		privateLocale := Locale("en-x-twemoji") // Private use subtag

		if string(privateLocale) != "en-x-twemoji" {
			t.Errorf("Private use locale string not preserved")
		}

		if privateLocale.Language() != "en" {
			t.Errorf("Private use locale language = %q, want en", privateLocale.Language())
		}
	})
}

func TestLocale_Performance(t *testing.T) {
	t.Run("String conversion performance", func(t *testing.T) {
		locale := ChineseSimplified
		iterations := 10000

		// Test that String() method is efficient
		for range iterations {
			_ = locale.String()
		}
		// If we reach here without timeout, the performance is acceptable
	})

	t.Run("Comparison performance", func(t *testing.T) {
		locales := []Locale{English, Chinese, Spanish, French, German}
		testLocale := Locale("zh-CN")
		iterations := 10000

		// Test that Contains() method is efficient
		for range iterations {
			for _, loc := range locales {
				_ = loc.Contains(testLocale)
			}
		}
		// If we reach here without timeout, the performance is acceptable
	})
}

func TestLocale_Compare(t *testing.T) {
	tests := []struct {
		name     string
		locale1  Locale
		locale2  Locale
		expected int
	}{
		// 完全匹配
		{"Exact match", English, English, 0},
		{"Exact match complex", Locale("zh-Hans-CN"), Locale("zh-Hans-CN"), 0},
		{"Empty both", Locale(""), Locale(""), 0},

		// 层次化比较：zh-Hans-CN → zh-Hans → zh-CN → zh (具体到通用)
		{"Base to script", Locale("zh"), Locale("zh-Hans"), 1},            // zh 比 zh-Hans 更通用
		{"Script to base", Locale("zh-Hans"), Locale("zh"), -1},           // zh-Hans 比 zh 更具体
		{"Script to region", Locale("zh-Hans"), Locale("zh-Hans-CN"), 1},  // zh-Hans 比 zh-Hans-CN 更通用
		{"Region to script", Locale("zh-Hans-CN"), Locale("zh-Hans"), -1}, // zh-Hans-CN 比 zh-Hans 更具体
		{"Base to region", Locale("zh"), Locale("zh-Hans-CN"), 1},         // zh 比 zh-Hans-CN 更通用
		{"Region to base", Locale("zh-Hans-CN"), Locale("zh"), -1},        // zh-Hans-CN 比 zh 更具体

		// 中文层次结构测试
		{"zh vs zh-Hans", Locale("zh"), Locale("zh-Hans"), 1},
		{"zh-Hans vs zh", Locale("zh-Hans"), Locale("zh"), -1},
		{"zh vs zh-Hant", Locale("zh"), Locale("zh-Hant"), 1},
		{"zh-Hant vs zh", Locale("zh-Hant"), Locale("zh"), -1},
		{"zh vs zh-CN", Locale("zh"), Locale("zh-CN"), 1},
		{"zh-CN vs zh", Locale("zh-CN"), Locale("zh"), -1},
		{"zh-Hans vs zh-Hans-CN", Locale("zh-Hans"), Locale("zh-Hans-CN"), 1},
		{"zh-Hans-CN vs zh-Hans", Locale("zh-Hans-CN"), Locale("zh-Hans"), -1},

		// 地区优先于脚本的规则（适用于所有语言）
		{"zh-CN vs zh-Hans", Locale("zh-CN"), Locale("zh-Hans"), -1},
		{"zh-Hans vs zh-CN", Locale("zh-Hans"), Locale("zh-CN"), 1},
		{"en-US vs en-Latn", Locale("en-US"), Locale("en-Latn"), -1},
		{"en-Latn vs en-US", Locale("en-Latn"), Locale("en-US"), 1},
		{"fr-FR vs fr-Latn", Locale("fr-FR"), Locale("fr-Latn"), -1},
		{"fr-Latn vs fr-FR", Locale("fr-Latn"), Locale("fr-FR"), 1},

		// 脚本不同的情况（按字典序）
		{"Script different Hans < Hant", Locale("zh-Hans"), Locale("zh-Hant"), -1},
		{"Script different Hant > Hans", Locale("zh-Hant"), Locale("zh-Hans"), 1},

		// 地区不同的情况（按字典序）
		{"Region different CN < TW", Locale("zh-Hans-CN"), Locale("zh-Hans-TW"), -1},
		{"Region different TW > CN", Locale("zh-Hans-TW"), Locale("zh-Hans-CN"), 1},
		{"Region different US > GB", Locale("en-US"), Locale("en-GB"), 1},
		{"Region different GB < US", Locale("en-GB"), Locale("en-US"), -1},

		// 英语层次结构测试
		{"en vs en-US", Locale("en"), Locale("en-US"), 1},
		{"en-US vs en", Locale("en-US"), Locale("en"), -1},
		{"en vs en-GB", Locale("en"), Locale("en-GB"), 1},
		{"en-GB vs en", Locale("en-GB"), Locale("en"), -1},

		// 语言不同（按字典序）
		{"Different languages en < zh", Locale("en"), Locale("zh"), -1},
		{"Different languages zh > en", Locale("zh"), Locale("en"), 1},

		// 空值处理
		{"Empty vs non-empty", Locale(""), English, -1},
		{"Non-empty vs empty", English, Locale(""), 1},

		// 复杂情况：相同具体程度，按字典序
		{"Same specificity zh-CN vs zh-TW", Locale("zh-CN"), Locale("zh-TW"), -1}, // CN < TW
		{"Same specificity zh-TW vs zh-CN", Locale("zh-TW"), Locale("zh-CN"), 1},  // TW > CN

		// 跨语言比较
		{"en vs zh", Locale("en"), Locale("zh"), -1}, // en < zh 字典序
		{"zh vs en", Locale("zh"), Locale("en"), 1},  // zh > en 字典序
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.locale1.Compare(tt.locale2)
			if result != tt.expected {
				t.Errorf("Locale(%q).Compare(%q) = %d, want %d",
					string(tt.locale1), string(tt.locale2), result, tt.expected)
			}
		})
	}
}

// ExampleLocale 展示优化后的 Locale 功能
func ExampleLocale() {
	// 简单的 locale 解析
	locale := Locale("zh-Hans-CN")

	// 获取语言、脚本、地区
	fmt.Printf("Language: %s, Script: %s, Region: %s\n",
		locale.Language(), locale.Script(), locale.Region())

	// 层次化匹配
	fmt.Println("zh contains zh-Hans:", Chinese.Contains(ChineseSimplified))

	// 比较 Locale（标准比较接口）
	fmt.Println("Compare zh-Hans-CN vs zh-Hans:", ChineseCN.Compare(ChineseSimplified))
	fmt.Println("Compare zh vs zh-Hans-CN:", Chinese.Compare(ChineseCN))
	fmt.Println("Compare zh-CN vs zh-Hans:", Locale("zh-CN").Compare(ChineseSimplified))
	fmt.Println("Compare en-US vs en-Latn:", Locale("en-US").Compare(Locale("en-Latn")))

	// Output:
	// Language: zh, Script: Hans, Region: CN
	// zh contains zh-Hans: true
	// Compare zh-Hans-CN vs zh-Hans: -1
	// Compare zh vs zh-Hans-CN: 1
	// Compare zh-CN vs zh-Hans: -1
	// Compare en-US vs en-Latn: -1
}
