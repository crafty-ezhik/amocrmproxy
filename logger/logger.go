package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

// NewLogger - возвращает указатель на экземпляр zap.Logger. Логгер выводит сообщения в файл и консоль
// Изменение уровня логирования через параметр debug.
// Также по-умолчанию включена ротация логов:
//   - размер 10Мб,
//   - хранение 14 дней,
//   - хранит по 3 файла, далее перезапись
func NewLogger(debug bool) *zap.Logger {
	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}

	// Настройка вывода в консоль
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	// Настройка вывода в файл с ротацией
	fileWriter := &lumberjack.Logger{
		Filename:   "./log/app.log",
		MaxSize:    10, // мегабайт
		MaxBackups: 3,
		MaxAge:     14, // дней
		Compress:   true,
	}

	// Формат JSON для файла
	fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	// Создаем кор логгера с разными уровнями и выводами
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)

	// Создаем логгер
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.WarnLevel))

	return logger
}
