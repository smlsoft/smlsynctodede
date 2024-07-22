package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"smlsynctodede/config"
	"strings"

	"gopkg.in/yaml.v2"
)

func FindConfig() (string, error) {
	// ลำดับการค้นหา config.yaml
	searchPaths := []string{
		"config.yaml",       // ในโฟลเดอร์ปัจจุบัน
		"../config.yaml",    // ในโฟลเดอร์แม่
		"build/config.yaml", // ในโฟลเดอร์ build
		filepath.Join("..", "build", "config.yaml"), // ในโฟลเดอร์ build ที่อยู่ในโฟลเดอร์แม่
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath, nil
		}
	}

	return "", fmt.Errorf("config.yaml not found in any of the search paths")
}

func GetUserInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func UpdateConfig(configPath string) error {
	var cfg config.Config

	// อ่านไฟล์ config เดิม
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return err
	}

	// รับข้อมูลใหม่จากผู้ใช้
	cfg.Database.Host = GetUserInput("Enter database host: ")
	cfg.Database.Port = GetIntInput("Enter database port: ")
	cfg.Database.User = GetUserInput("Enter database user: ")
	cfg.Database.Password = GetUserInput("Enter database password: ")

	// รับชื่อฐานข้อมูล
	var databases []config.DatabaseConfig
	for {
		dbName := GetUserInput("Enter database name (or press Enter to finish): ")
		if dbName == "" {
			break
		}
		databases = append(databases, config.DatabaseConfig{Name: dbName})
	}
	cfg.Databases = databases

	cfg.API.Key = GetUserInput("Enter API key In Merchant: ")

	// เขียนข้อมูลลงไฟล์ config
	newContent, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, newContent, 0644)
}
