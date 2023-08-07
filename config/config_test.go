package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	// Сохраняем старые аргументы командной строки и восстанавливаем их после теста
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Устанавливаем фиктивные аргументы командной строки для теста
	os.Args = []string{"test", "-a", "testaddress", "-b", "testurl"}

	// Вызываем функцию ParseFlags() и проверяем, что получили ожидаемые значения
	cfg := ParseFlags()

	assert.Equal(t, "testaddress", cfg.ServerAddress)
	assert.Equal(t, "testurl", cfg.BaseURL)
}
