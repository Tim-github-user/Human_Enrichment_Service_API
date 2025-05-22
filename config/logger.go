package config

import (
	"os"
	"github.com/sirupsen/logrus"
)

// Log - это глобальный экземпляр логгера
var Log *logrus.Logger

func InitLogger() {
	Log = logrus.New() // Создаем новый экземпляр логгера.

	// Устанавливаем форматтер для логов. TextFormatter делает логи читаемыми.
	// FullTimestamp: true добавляет полную метку времени к каждой записи.
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Устанавливаем вывод логов в стандартный вывод (консоль).
	Log.SetOutput(os.Stdout)

	// Устанавливаем уровень логирования.
	// logrus.DebugLevel: Будут выводиться все логи (Debug, Info, Warn, Error, Fatal, Panic).
	Log.SetLevel(logrus.DebugLevel)
}