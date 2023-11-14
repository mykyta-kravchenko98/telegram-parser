package config

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Google   GoogleCredentials   `yaml:"google"`
	Telegram TelegramCredentials `yaml:"telegram"`
}

type TelegramCredentials struct {
	Token   string `yaml:"token"`
	AdminId int64  `yaml:"adminId"`
}

type GoogleCredentials struct {
	Type                    string `yaml:"type"`
	ProjectID               string `yaml:"project_id"`
	PrivateKeyID            string `yaml:"private_key_id"`
	PrivateKey              string `yaml:"private_key"`
	ClientEmail             string `yaml:"client_email"`
	ClientID                string `yaml:"client_id"`
	AuthURI                 string `yaml:"auth_uri"`
	TokenURI                string `yaml:"token_uri"`
	AuthProviderX509CertURL string `yaml:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `yaml:"client_x509_cert_url"`
	UniverseDomain          string `yaml:"universe_domain"`
	SpreadsheetID           string `yaml:"spreadsheet_id"`
}

var (
	vp     *viper.Viper
	config *Config
)

// LoadConfigJSON is a init method that find config json file and initialize Config struct. Must be called in main.go. Using in local env
func LoadConfigJSON(env string) (*Config, error) {
	vp = viper.New()

	vp.SetConfigType("json")
	vp.SetConfigName(env)
	vp.AddConfigPath("../config/")
	vp.AddConfigPath("../../config/")
	vp.AddConfigPath("config/")
	vp.AddConfigPath("app/")

	err := vp.ReadInConfig()
	if err != nil {
		return &Config{}, err
	}

	// Перезапись значений из переменных окружения
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	adminID := os.Getenv("TELEGRAM_BOT_ADMIN_ID")

	if telegramToken != "" {
		vp.Set("telegram.token", telegramToken)
	}
	if adminID != "" {
		vp.Set("telegram.adminId", adminID)
	}

	err = vp.Unmarshal(&config)
	if err != nil {
		return &Config{}, err
	}

	return config, err
}

// LoadConfigYAML is a init method that find config yaml file and initialize Config struct. Using when deploying on prod
func LoadConfigYAML() (*Config, error) {
	var filePath string
	env := os.Getenv("ENVIRONMENT")
	fmt.Println("ENVIRONMENT:", env)

	// Проверяем переменную окружения, чтобы определить путь к файлу конфигурации
	if env == "production" {
		filePath = "/app/config/config.yml"
	} else {
		filePath = "config/config.yml"
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening config file: %s, error: %v", filePath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Parse the YAML data into the Config struct
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, err
}

// GetConfig method provide geting already init config data
func GetConfig() *Config {
	return config
}
