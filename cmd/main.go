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
	"sync"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	configPath, err := utils.FindConfig()
	if err != nil {
		log.Fatalf("Error finding config: %v", err)
	}

	if err := config.LoadConfig(configPath); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err := database.TestPostgresConnection(config.AppConfig); err != nil {
		log.Fatalf("Database connection test failed: %v", err)
	}

	logPath := filepath.Join(filepath.Dir(configPath), "sync_log.txt")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	fmt.Printf("Starting synchronization process. Logs saved to %s\n", logPath)

	timeStart := time.Now()
	log.Printf("%s=== Start synchronization process ===%s", logging.ColorGreen, logging.ColorReset)
	log.Printf("%sStart Time: %s%s", logging.ColorCyan, timeStart.Format("2006-01-02 15:04:05.000"), logging.ColorReset)

	logging.InitResults()

	syncFunctions := []struct {
		name     string
		function func(models.DatabaseModel, string) error
	}{
		{"ap_supplier", functions.SyncApSupplierToMongoDB},
		{"ar_customer", functions.SyncArCustomerToMongoDB},
		{"ic_inventory", functions.SyncIcInventoryToMongoDB},
		{"ic_unit", functions.SyncIcUnitToMongoDB},
	}

	for _, dbName := range config.GetDatabaseList() {
		var wg sync.WaitGroup
		errChan := make(chan error, len(syncFunctions))

		for _, syncFunc := range syncFunctions {
			wg.Add(1)
			go func(sf struct {
				name     string
				function func(models.DatabaseModel, string) error
			}, db string) {
				defer wg.Done()
				if err := sf.function(models.DatabaseModel{DatabaseName: db}, config.AppConfig.API.Key); err != nil {
					log.Printf("%sError in %s sync: %v%s", logging.ColorRed, sf.name, err, logging.ColorReset)
					errChan <- err
				}
			}(syncFunc, dbName)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			if err != nil {
				log.Printf("%sSync completed with errors for database %s%s", logging.ColorYellow, dbName, logging.ColorReset)
				break
			}
		}
	}

	timeStop := time.Now()
	log.Printf("%s=== Synchronization process completed ===%s", logging.ColorGreen, logging.ColorReset)
	log.Printf("%sTotal Time: %.3f seconds%s", logging.ColorYellow, timeStop.Sub(timeStart).Seconds(), logging.ColorReset)

	logging.PrintSummary(timeStart, timeStop)

	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}
