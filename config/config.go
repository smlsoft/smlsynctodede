package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
	API struct {
		Key     string `yaml:"key"`
		BaseURL string `yaml:"base_url"`
	} `yaml:"api"`
	Databases []DatabaseConfig `yaml:"databases"`
}

type DatabaseConfig struct {
	Name string `yaml:"name"`
}

type PartService struct {
	ServiceName string `yaml:"service_name"`
	PartName    string `yaml:"part_name"`
}

var AppConfig Config

var PartServices = []PartService{
	{ServiceName: "creditor", PartName: "debtaccount/creditor/bulk"},
	{ServiceName: "debtor", PartName: "debtaccount/debtor/bulk"},
	{ServiceName: "productbarcode", PartName: "product/barcode/import"},
	{ServiceName: "unit", PartName: "unit/bulk"},
}

func LoadConfig(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot read config file: %w", err)
	}

	if err := yaml.Unmarshal(content, &AppConfig); err != nil {
		return fmt.Errorf("cannot parse config file: %w", err)
	}

	return validateConfig(AppConfig)
}

func validateConfig(cfg Config) error {
	if cfg.Database.Host == "" {
		return fmt.Errorf("config: database.host is required")
	}
	if cfg.Database.Port == 0 {
		return fmt.Errorf("config: database.port is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("config: database.user is required")
	}
	if cfg.API.Key == "" {
		return fmt.Errorf("config: api.key is required")
	}
	if cfg.API.BaseURL == "" {
		return fmt.Errorf("config: api.base_url is required")
	}
	if len(cfg.Databases) == 0 {
		return fmt.Errorf("config: at least one database must be configured")
	}
	return nil
}

func GetDatabaseList() []string {
	names := make([]string, len(AppConfig.Databases))
	for i, db := range AppConfig.Databases {
		names[i] = db.Name
	}
	return names
}
