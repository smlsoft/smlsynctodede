package config

import (
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

var AppConfig Config

func LoadConfig(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(content, &AppConfig)
	if err != nil {
		return err
	}

	return nil
}

func GetDatabaseList() []string {
	var dbList []string
	for _, db := range AppConfig.Databases {
		dbList = append(dbList, db.Name)
	}
	return dbList
}
