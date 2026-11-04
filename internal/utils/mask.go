package utils

import (
	"strings"
	"unicode/utf8"
)

// MaskName che bớt các ký tự trong tên để bảo vệ danh tính khách hàng
// Ví dụ: "Nguyễn Đình Thiện" -> "N***n Đ**h T***n"
func MaskName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}

	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return ""
	}

	maskedWords := make([]string, len(words))
	for i, word := range words {
		runes := []rune(word)
		n := len(runes)
		if n <= 1 {
			maskedWords[i] = "*"
		} else if n == 2 {
			maskedWords[i] = string(runes[0]) + "*"
		} else if n == 3 {
			maskedWords[i] = string(runes[0]) + "*" + string(runes[2])
		} else {
			maskedWords[i] = string(runes[0]) + strings.Repeat("*", n-2) + string(runes[n-1])
		}
	}

	return strings.Join(maskedWords, " ")
}

// MaskPhone che các ký tự ở giữa của số điện thoại
// Ví dụ: "0988111222" -> "098***1222"
func MaskPhone(phone string) string {
	cleaned := strings.TrimSpace(phone)
	length := utf8.RuneCountInString(cleaned)
	if length <= 6 {
		return cleaned
	}

	runes := []rune(cleaned)
	prefix := string(runes[:3])
	suffix := string(runes[length-3:])
	return prefix + strings.Repeat("*", length-6) + suffix
}

// MaskAddress che số nhà hoặc thông tin ngõ hẻm chi tiết, chỉ để lại tuyến đường/khu vực
// Ví dụ: "123 Tô Hiến Thành, Nha Trang" -> "*** Tô Hiến Thành, Nha Trang"
func MaskAddress(address string) string {
	trimmed := strings.TrimSpace(address)
	if trimmed == "" {
		return ""
	}

	parts := strings.Split(trimmed, ",")
	if len(parts) > 1 {
		// Che phần đầu tiên (thường là số nhà/ngõ hẻm cụ thể)
		firstPartRunes := []rune(strings.TrimSpace(parts[0]))
		if len(firstPartRunes) > 0 {
			parts[0] = "***"
		}
		return strings.Join(parts, ",")
	}

	// Nếu không có dấu phẩy, che 40% ký tự đầu
	runes := []rune(trimmed)
	cut := len(runes) / 2
	if cut < 3 {
		cut = 3
	}
	if cut >= len(runes) {
		return "***"
	}
	return "*** " + string(runes[cut:])
}
