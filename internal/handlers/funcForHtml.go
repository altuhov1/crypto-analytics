package handlers

import (
	"strconv"
	"strings"
)

func formatNumber(num int64) string {
	s := strconv.FormatInt(num, 10)
	parts := []string{}

	// Разбиваем число на группы по 3 цифры
	for i := len(s); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{s[start:i]}, parts...)
	}

	return strings.Join(parts, ",")
}

// add функция для сложения чисел в шаблоне
func add(a, b int) int {
	return a + b
}
