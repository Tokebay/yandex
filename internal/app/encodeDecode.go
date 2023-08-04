package app

import "strings"

// Функция для кодирования числа в base62
func encodeBase62(n int64) string {
	alphabet := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base := int64(len(alphabet))
	result := ""

	for n > 0 {
		remainder := n % base
		result = string(alphabet[remainder]) + result
		n /= base
	}

	return result
}

// Функция для декодирования строки из base62 обратно в число
func decodeBase62(s string) int64 {
	alphabet := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base := int64(len(alphabet))
	result := int64(0)

	for _, c := range s {
		result = result*base + int64(strings.IndexRune(alphabet, c))
	}

	return result
}
