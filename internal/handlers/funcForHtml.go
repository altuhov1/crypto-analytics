package handlers

import (
	"fmt"
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

func formatMoney(amount float64) string {
	// Форматируем с разделителями тысяч
	str := fmt.Sprintf("%.0f", amount)
	n := len(str)
	if n <= 3 {
		return "$" + str
	}

	var result strings.Builder
	for i, char := range str {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(char)
	}
	return "$" + result.String()
}
