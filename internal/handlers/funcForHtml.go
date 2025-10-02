package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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

func parseTime(timeStr string) *time.Time {
	if timeStr == "" {
		return nil
	}

	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC3339,
		"Mon, 2 Jan 2006 15:04:05 MST",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"02 Jan 2006 15:04:05 MST",
		"2006-01-02 15:04:05 -0700",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return &t
		}
	}
	return nil
}
