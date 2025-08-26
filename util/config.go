package util

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	onceCFG sync.Once
	config  *Config
	cfgPath = "configuration/config.yaml"
)

// ConfigTask представляет конфигурацию приложения требуемое в задание.
type ConfigTask struct {
	ServerAddres    string `json:"server_address"`    // аналог переменной окружения SERVER_ADDRESS или флага -a
	FileStoragePath string `json:"file_storage_path"` // аналог переменной окружения FILE_STORAGE_PATH или флага -f
}

// Config представляет конфигурацию приложения.
// Содержит настройки для логирования, сервера и аутентификации.
type Config struct {
	Logger          LoggerCfg `yaml:"Logger"`
	Server          Server    `yaml:"Server"`
	Auth            Auth      `yaml:"Auth"`
	FileStoragePath string    `yaml:"FileStoragePath"`
}

// Auth содержит конфигурацию, связанную с аутентификацией.
type Auth struct {
	SecretKey  string `yaml:"SecretKey"`
	CookieName string `yaml:"CookieName"`
}

// Server содержит конфигурацию HTTP-сервера.
type Server struct {
	ServerAddress string `yaml:"-"`
	Address       string `yaml:"Address"`
	RTimeout      int64  `yaml:"RTimeout"`
	WTimeout      int64  `yaml:"WTimeout"`
	Port          uint   `yaml:"Port"`
}

func parseConfig(st interface{}, cfgPath string) {
	f, err := os.Open(cfgPath)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "error occurred while opening cfg file"))
	}

	fi, err := f.Stat()
	if err != nil {
		log.Fatal(errors.WithMessage(err, "error occurred while getting file stats"))
	}

	data := make([]byte, fi.Size())
	_, err = f.Read(data)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "error occurred while reading data"))
	}

	err = yaml.Unmarshal(data, st)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "error occurred while unmashaling data"))
	}
}

func parseConfigTask(st interface{}, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.WithMessage(err, "error occurred while opening JSON config file")
	}
	if err := json.Unmarshal(data, st); err != nil {
		return errors.WithMessage(err, "error occurred while unmarshaling JSON config")
	}
	return nil
}

// GetConfig возвращает указатель на глобальную конфигурацию приложения.
// Инициализирует конфигурацию при первом вызове, загружая данные из файла и переменных окружения.
// Возвращает *Config.
func GetConfig() *Config {
	onceCFG.Do(func() {
		var (
			conf                           Config
			configTask                     ConfigTask
			configPath                     string
			serverAddress, fileStoragePath string
		)
		parseConfig(&conf, cfgPath)

		flag.StringVar(&configPath, "c", "", "Path to JSON config file")
		flag.StringVar(&configPath, "config", "", "Path to JSON config file (long)")
		flag.StringVar(&serverAddress, "a", "", "HTTP server address")
		flag.StringVar(&fileStoragePath, "f", "", "Path to file storage")
		flag.Parse()

		if envConfig, exists := os.LookupEnv("CONFIG"); exists {
			configPath = envConfig
		}

		if configPath != "" {
			if err := parseConfigTask(&configTask, configPath); err != nil {
				log.Fatalln("error occurred while parsing config task ", err)
			}
		}

		if envAddr, exists := os.LookupEnv("SERVER_ADDRESS"); exists {
			conf.Server.ServerAddress = envAddr
		} else if serverAddress != "" {
			conf.Server.ServerAddress = serverAddress
		} else if configTask.ServerAddres != "" {
			conf.Server.ServerAddress = configTask.ServerAddres
		} else {
			conf.Server.ServerAddress = fmt.Sprintf("%s:%d", conf.Server.Address, conf.Server.Port)
		}

		if envPath, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
			conf.FileStoragePath = envPath
		} else if fileStoragePath != "" {
			conf.FileStoragePath = fileStoragePath
		} else if configTask.FileStoragePath != "" {
			conf.FileStoragePath = configTask.FileStoragePath
		}

		config = &conf
	})

	if config == nil {
		log.Fatal("nil config")
	}

	return config
}
