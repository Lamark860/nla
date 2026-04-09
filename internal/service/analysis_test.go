package service

import (
	"testing"
)

func TestParseRatingFromResponse(t *testing.T) {
	tests := []struct {
		name string
		text string
		want *int
	}{
		// [RATING:XX] — highest priority
		{"rating tag", "Вот анализ облигации.\n[RATING:72]", intPtr(72)},
		{"rating tag with spaces", "[RATING: 85 ]", intPtr(85)},
		{"rating tag zero", "[RATING:0]", intPtr(0)},
		{"rating tag 100", "[RATING:100]", intPtr(100)},

		// Итоговая оценка XX/100
		{"itog score /100", "Итоговая оценка: **72 / 100**", intPtr(72)},
		{"itog score compact", "Итоговая оценка (по 100-балльной шкале): **73/100**", intPtr(73)},
		{"itog score range", "Итоговая оценка (0–100): **74/100**", intPtr(74)},

		// Итоговая оценка XX баллов
		{"itog ballов", "Итоговая оценка: примерно 65 баллов из 100.", intPtr(65)},
		{"itog balla", "Итоговая оценка: 58 балла", intPtr(58)},

		// Оценка XX/100
		{"ocenka /100", "Оценка: 81/100", intPtr(81)},

		// Оценка XX баллов
		{"ocenka ball", "Оценка: 55 баллов", intPtr(55)},

		// Bold **XX/100**
		{"bold rating", "Результат: **67 / 100**", intPtr(67)},
		{"bold compact", "Итого **88/100**", intPtr(88)},

		// Plain XX/100 (last occurrence)
		{"plain /100 last", "Блок 1: 35/45\nБлок 2: 20/25\nИтого: 72/100", intPtr(72)},

		// No rating
		{"no rating", "Облигация имеет хорошие характеристики.", nil},
		{"empty", "", nil},

		// Out of range
		{"rating over 100", "[RATING:150]", nil},

		// Priority: [RATING] wins over everything
		{"priority", "Итоговая оценка: 60/100\n[RATING:72]", intPtr(72)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseRatingFromResponse(tt.text)
			if tt.want == nil {
				if got != nil {
					t.Errorf("got %d, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Errorf("got nil, want %d", *tt.want)
				return
			}
			if *got != *tt.want {
				t.Errorf("got %d, want %d", *got, *tt.want)
			}
		})
	}
}

func intPtr(v int) *int {
	return &v
}
