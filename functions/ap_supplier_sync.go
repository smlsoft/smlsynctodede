package functions

import (
	"database/sql"
	"fmt"
	"log"
	"smlsynctodede/models"
	"smlsynctodede/myglobal"
	"time"
)

func SyncApSupplierToMongoDB(database models.DatabaseModel, apiKey string) error {
	tableName := "ap_supplier"
	start := time.Now()
	log.Printf("▶ Start Sync Table %s%s%s: %s", myglobal.ColorYellow, tableName, myglobal.ColorReset, database.DatabaseName)

	// Step 1: Connect to PostgreSQL
	stepStart := time.Now()
	db, err := sql.Open("postgres", myglobal.GetPostgreSQLConnectionString(database.DatabaseName))
	if err != nil {
		myglobal.LogError(fmt.Sprintf("Error connecting to PostgreSQL for table %s", tableName), err)
		return err
	}
	defer db.Close()
	log.Printf("Step 1: Connected to PostgreSQL (%.2f seconds)", time.Since(stepStart).Seconds())

	// Step 2: Query data from PostgreSQL
	stepStart = time.Now()
	rows, err := db.Query(`
        SELECT 
            code, 
            name_1
        FROM 
            ap_supplier LIMIT 100
    `)
	if err != nil {
		myglobal.LogError(fmt.Sprintf("Error querying PostgreSQL for table %s", tableName), err)
		return err
	}
	defer rows.Close()
	log.Printf("Step 2: Queried data from PostgreSQL (%.2f seconds)", time.Since(stepStart).Seconds())

	// Step 3: Process query results
	stepStart = time.Now()
	var creditors []models.MongoCreditorModel
	for rows.Next() {
		var code, name sql.NullString
		err := rows.Scan(&code, &name)
		if err != nil {
			log.Printf("Error scanning row for table %s: %v", tableName, err)
			continue
		}

		creditor := models.MongoCreditorModel{
			Code: code.String,
			Names: []models.LanguageNameModel{
				{
					Code:     "th",
					Name:     name.String,
					Isauto:   false,
					Isdelete: false,
				},
			},
		}

		creditors = append(creditors, creditor)
	}
	log.Printf("Step 3: Processed %d items (%.2f seconds)", len(creditors), time.Since(stepStart).Seconds())

	// Step 4: Send data to API
	stepStart = time.Now()
	batchSize := 50
	totalItems := len(creditors)
	for i := 0; i < totalItems; i += batchSize {
		batchStart := time.Now() // เพิ่มจุดเริ่มต้นในการจับเวลา
		end := i + batchSize
		if end > totalItems {
			end = totalItems
		}
		batch := creditors[i:end]

		responseBody, err := myglobal.SendDataToAPI("creditor", apiKey, batch)
		if err != nil {
			if err.Error() == "API request failed with status code 401: {\"message\":\"Token Invalid.\",\"success\":false}" {
				myglobal.LogError("Authentication failed. Please check your API key.", fmt.Errorf(string(responseBody)))
				return err // Return error to stop the synchronization process
			}
			myglobal.LogError(fmt.Sprintf("Error sending data to API for table %s (batch %d-%d)", tableName, i+1, end), err)
			return err // Return error to stop the synchronization process
		}

		batchDuration := time.Since(batchStart) // เพิ่มจุดสิ้นสุดในการจับเวลา
		log.Printf("Sent batch %d-%d of %d for table %s (%.2f seconds)", i+1, end, totalItems, tableName, batchDuration.Seconds())
	}
	log.Printf("Step 4: Sent all data to API (%.2f seconds)", time.Since(stepStart).Seconds())

	duration := time.Since(start)
	myglobal.LogSuccess(fmt.Sprintf("Sync Table %s", tableName), database.DatabaseName, duration, totalItems)
	myglobal.LogResult(database.DatabaseName, fmt.Sprintf("Sync %s ToMongoDB", tableName), duration, totalItems)

	return nil
}
