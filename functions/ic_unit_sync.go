package functions

import (
	"database/sql"
	"fmt"
	"log"
	"smlsynctodede/database"
	"smlsynctodede/logging"
	"smlsynctodede/models"
	"smlsynctodede/utils"
	"time"
)

func SyncIcUnitToMongoDB(databases models.DatabaseModel, apiKey string) error {
	tableName := "ic_unit"
	start := time.Now()
	logging.LogStartSync(tableName, databases.DatabaseName)

	// Step 1: Connect to PostgreSQL
	stepStart := time.Now()
	db, err := sql.Open("postgres", database.GetPostgreSQLConnectionString(databases.DatabaseName))
	if err != nil {
		logging.LogError(fmt.Sprintf("Error connecting to PostgreSQL for table %s", tableName), err)
		return err
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		logging.LogError(fmt.Sprintf("Error pinging PostgreSQL for table %s", tableName), err)
		return err
	}
	log.Printf("Step 1: Connected to PostgreSQL (%.2f seconds)", time.Since(stepStart).Seconds())

	// Step 2: Query data from PostgreSQL
	stepStart = time.Now()
	rows, err := db.Query(`SELECT code, name_1 FROM ic_unit`)
	if err != nil {
		logging.LogError(fmt.Sprintf("Error querying PostgreSQL for table %s", tableName), err)
		return err
	}
	defer rows.Close()
	log.Printf("Step 2: Queried data from PostgreSQL (%.2f seconds)", time.Since(stepStart).Seconds())

	// Step 3: Process query results
	stepStart = time.Now()
	var units []models.MongoUnitModel
	for rows.Next() {
		var code, name sql.NullString
		if err := rows.Scan(&code, &name); err != nil {
			log.Printf("Error scanning row for table %s: %v", tableName, err)
			continue
		}
		units = append(units, models.MongoUnitModel{
			UnitCode: code.String,
			Names: []models.LanguageNameModel{
				{Code: "th", Name: name.String, Isauto: false, Isdelete: false},
			},
		})
	}
	if err := rows.Err(); err != nil {
		logging.LogError(fmt.Sprintf("Error iterating rows for table %s", tableName), err)
		return err
	}
	log.Printf("Step 3: Processed %d items (%.2f seconds)", len(units), time.Since(stepStart).Seconds())

	// Step 4: Send data to API
	stepStart = time.Now()
	batchSize := 50
	totalItems := len(units)
	for i := 0; i < totalItems; i += batchSize {
		end := i + batchSize
		if end > totalItems {
			end = totalItems
		}
		batchStart := time.Now()

		_, err := utils.SendDataToAPI("unit", apiKey, units[i:end])
		if err != nil {
			logging.LogError(fmt.Sprintf("Error sending batch %d-%d for table %s", i+1, end, tableName), err)
			return err
		}
		log.Printf("Sent batch %d-%d of %d for table %s (%.2f seconds)", i+1, end, totalItems, tableName, time.Since(batchStart).Seconds())
	}
	log.Printf("Step 4: Sent all data to API (%.2f seconds)", time.Since(stepStart).Seconds())

	duration := time.Since(start)
	logging.LogSuccess(fmt.Sprintf("Successful Table %s", tableName), databases.DatabaseName, duration, totalItems)
	logging.LogResult(databases.DatabaseName, fmt.Sprintf("Sync %s ToMongoDB", tableName), duration, totalItems)

	return nil
}
