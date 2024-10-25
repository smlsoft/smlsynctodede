package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"smlsynctodede/config"
	"smlsynctodede/database"
	"smlsynctodede/functions"
	"smlsynctodede/logging"
	"smlsynctodede/models"
	"smlsynctodede/utils"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	configPath, err := utils.FindConfig()
	if err != nil {
		log.Fatalf("Error finding config: %v", err)
	}

	/// ใช้ป้อน config
	// err = utils.UpdateConfig(configPath)
	// if err != nil {
	// 	log.Fatalf("Error updating config: %v", err)
	// }

	err = config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	err = database.TestPostgresConnection(config.AppConfig)
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
	log.Printf("%s=== Start synchronization process ===%s", logging.ColorGreen, logging.ColorReset)
	log.Printf("%sStart Time: %s%s%s", logging.ColorCyan, timeStart.Format("2006-01-02 15:04:05.000"), logging.ColorCyan, logging.ColorReset)

	logging.InitResults()

	stopChan := make(chan bool)
	go utils.StartLoadingAnimation("Sync data", stopChan, ": ")

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
				log.Printf("%sError in %s sync: %v%s", logging.ColorRed, syncFunc.name, err, logging.ColorReset)
				// Optionally, you can break here if you want to stop on first error
				// break
			}
		}
	}

	stopChan <- true // Stop the loading animation

	timeStop := time.Now()
	log.Printf("%s=== Synchronization process completed ===%s", logging.ColorGreen, logging.ColorReset)
	log.Printf("%sTotal Time: %s%.3f seconds%s", logging.ColorYellow, logging.ColorReset, timeStop.Sub(timeStart).Seconds(), logging.ColorReset)

	logging.PrintSummary(timeStart, timeStop)

	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}
