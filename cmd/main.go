package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"smlsynctodede/config"
	"smlsynctodede/functions"
	"smlsynctodede/models"
	"smlsynctodede/myglobal"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

func findConfig() (string, error) {
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

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func updateConfig(configPath string) error {
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
	cfg.Database.Host = getUserInput("Enter database host: ")
	cfg.Database.Port = myglobal.GetIntInput("Enter database port: ")
	cfg.Database.User = getUserInput("Enter database user: ")
	cfg.Database.Password = getUserInput("Enter database password: ")

	// รับชื่อฐานข้อมูล
	var databases []config.DatabaseConfig
	for {
		dbName := getUserInput("Enter database name (or press Enter to finish): ")
		if dbName == "" {
			break
		}
		databases = append(databases, config.DatabaseConfig{Name: dbName})
	}
	cfg.Databases = databases

	cfg.API.Key = getUserInput("Enter API key In Merchant: ")

	// เขียนข้อมูลลงไฟล์ config
	newContent, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, newContent, 0644)
}

func testPostgresConnection(cfg config.Config) error {
	for _, db := range cfg.Databases {
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, db.Name)

		testDB, err := sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("error opening connection to %s: %v", db.Name, err)
		}
		defer testDB.Close()

		err = testDB.Ping()
		if err != nil {
			return fmt.Errorf("error connecting to %s: %v", db.Name, err)
		}
		fmt.Printf("Successfully connected to database: %s\n", db.Name)
	}
	return nil
}

func main() {
	configPath, err := findConfig()
	if err != nil {
		log.Fatalf("Error finding config: %v", err)
	}

	/// ใช้ป้อน config
	err = updateConfig(configPath)
	if err != nil {
		log.Fatalf("Error updating config: %v", err)
	}

	err = config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	err = testPostgresConnection(config.AppConfig)
	if err != nil {
		log.Fatalf("Database connection test failed: %v", err)
	}

	logPath := filepath.Join(filepath.Dir(configPath), "sync_log.txt")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	fmt.Println("Starting synchronization process. Please wait...")
	fmt.Printf("Logs will be displayed here and saved to %s\n", logPath)

	timeStart := time.Now()
	log.Printf("%s=== Start synchronization process ===%s", myglobal.ColorGreen, myglobal.ColorReset)
	log.Printf("%sStart Time: %s%s%s", myglobal.ColorCyan, timeStart.Format("2006-01-02 15:04:05.000"), myglobal.ColorCyan, myglobal.ColorReset)

	myglobal.InitResults()

	syncFunctions := []struct {
		name     string
		function func(models.DatabaseModel, string) error
	}{
		{"ap_supplier", functions.SyncApSupplierToMongoDB},
		{"ar_customer", functions.SyncArCustomerToMongoDB},
		{"ic_inventory", functions.SyncIcInventoryToMongoDB},
	}

	for _, dbName := range config.GetDatabaseList() {
		for _, syncFunc := range syncFunctions {
			err := syncFunc.function(models.DatabaseModel{DatabaseName: dbName}, config.AppConfig.API.Key)
			if err != nil {
				log.Printf("%sError in %s sync: %v%s", myglobal.ColorRed, syncFunc.name, err, myglobal.ColorReset)
				// Optionally, you can break here if you want to stop on first error
				// break
			}
		}
	}

	timeStop := time.Now()
	log.Printf("%s=== Synchronization process completed ===%s", myglobal.ColorGreen, myglobal.ColorReset)
	log.Printf("%sTotal Time: %s%.3f seconds%s", myglobal.ColorYellow, myglobal.ColorReset, timeStop.Sub(timeStart).Seconds(), myglobal.ColorReset)

	myglobal.PrintSummary(timeStart, timeStop)

	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}
