package util

import (
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// LoggerCfg представляет конфигурацию логгера.
// Содержит уровень логирования, файл и системный журнал.
type LoggerCfg struct {
	Level  log.Level `yaml:"Level"`
	File   LogFile   `yaml:"File"`
	SysLog SysLog    `yaml:"Syslog"`
}

// LogFile представляет конфигурацию файла для логирования.
// Содержит флаг включения, имя файла, максимальный размер файла, максимальное количество резервных копий и максимальный возраст файла.
type LogFile struct {
	Enabled    bool   `yaml:"Enabled"`
	FileName   string `yaml:"FileName"`
	MaxSize    int    `yaml:"MaxSize"`
	MaxBackups int    `yaml:"MaxBackups"`
	MaxAge     int    `yaml:"MaxAge"`
}

// SysLog представляет конфигурацию системного журнала.
// Содержит флаг включения, адрес, сеть и тег.
type SysLog struct {
	Enabled bool   `yaml:"Enabled"`
	Address string `yaml:"Address"`
	Network string `yaml:"Network"`
	Tag     string `yaml:"Tag"`
}

// logger представляет глобальный экземпляр логгера logrus.
var logger *log.Logger

// InitLogger инициализирует глобальный экземпляр логгера logrus.
// Принимает конфигурацию логгера.
// Инициализирует логгер с указанными параметрами.
func InitLogger(conf LoggerCfg) {
	logger = &log.Logger{
		Out:          os.Stdout,
		Hooks:        nil,
		Formatter:    &log.JSONFormatter{},
		ReportCaller: false,
		Level:        conf.Level,
		ExitFunc:     os.Exit,
	}
	if conf.File.Enabled {
		if len(conf.File.FileName) == 0 {
			panic(errors.New("parameter FileName not defined"))
		}
		rotateHook, err := NewRotateFileHook(RotateFileConfig{
			Filename:   conf.File.FileName,
			MaxSize:    conf.File.MaxSize,
			MaxBackups: conf.File.MaxBackups,
			MaxAge:     conf.File.MaxAge,
			Level:      conf.Level,
			Formatter:  &log.JSONFormatter{},
		})
		if err != nil {
			logger.Info("Failed to log to file")
			return
		}
		logger.Hooks = log.LevelHooks{}
		logger.AddHook(rotateHook)
	}
	if conf.SysLog.Enabled {
		syslog, err := NewSyslogHook(conf)
		if err != nil {
			logger.Info("Failed to log to syslog")
			return
		}
		if logger.Hooks == nil {
			logger.Hooks = log.LevelHooks{}
		}
		logger.AddHook(syslog)
	}
}

// GetLogger возвращает глобальный экземпляр логгера logrus.
func GetLogger() *log.Logger {
	return logger
}
