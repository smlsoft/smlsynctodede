package functions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"smlsynctodede/database"
	"smlsynctodede/logging"
	"smlsynctodede/models"
	"smlsynctodede/utils"
	"time"
)

func SyncIcInventoryToMongoDB(databases models.DatabaseModel, apiKey string) error {
	tableName := "ic_inventory"
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
	// Fixed CASE logic: OR -> AND to correctly fallback price_0 only when price=0 AND price_0 is non-empty/non-zero
	stepStart = time.Now()
	rows, err := db.Query(`
		SELECT code, barcode, unit_code, unit_name, item_type, drink_type, tax_type, have_point, ic_name, group_main, group_name,
		CASE WHEN price <> 0 THEN price
		WHEN COALESCE(price_0, '') <> '' AND price_0 <> '0' THEN CAST(COALESCE(NULLIF(price_0, ''), '0') AS numeric)
		ELSE 0 END AS price,
		price_member
		FROM (
			SELECT ic.code, ic_barcode.barcode, ic_barcode.unit_code AS unit_code,
			(SELECT name_1 FROM ic_unit WHERE ic_unit.code = ic_barcode.unit_code) AS unit_name,
			ic.item_type, ic.drink_type, ic.tax_type, ic_detail.have_point,
			ic.name_1 AS ic_name, ic.group_main, ic_group.name_1 AS group_name,
			ic_barcode.price, ic_barcode.price_member, ic_price.price_0
			FROM ic_inventory AS ic
			LEFT JOIN ic_inventory_detail AS ic_detail ON ic.code = ic_detail.ic_code
			LEFT JOIN ic_inventory_barcode AS ic_barcode ON ic.code = ic_barcode.ic_code
			LEFT JOIN ic_inventory_price_formula AS ic_price ON ic_price.ic_code = ic_barcode.ic_code AND ic_price.unit_code = ic_barcode.unit_code
			LEFT JOIN ic_unit_use AS ic_unit ON ic.code = ic_unit.ic_code AND ic_barcode.ic_code = ic_unit.ic_code AND ic_barcode.unit_code = ic_unit.code
			LEFT JOIN ic_unit AS unit ON ic.unit_standard = unit.code
			LEFT JOIN ic_group AS ic_group ON ic_group.code = ic.group_main
			WHERE COALESCE(ic_barcode.barcode, '') <> ''
			ORDER BY ic.code
		) AS temp1
	`)
	if err != nil {
		logging.LogError(fmt.Sprintf("Error querying PostgreSQL for table %s", tableName), err)
		return err
	}
	defer rows.Close()
	log.Printf("Step 2: Queried data from PostgreSQL (%.2f seconds)", time.Since(stepStart).Seconds())

	// Step 3: Process query results
	stepStart = time.Now()
	var inventories []models.MongoProductBarcodeModel
	for rows.Next() {
		var code, barcode, unitCode, unitName sql.NullString
		var itemType, drinkType, taxType sql.NullInt64
		var havePoint sql.NullBool
		var icName, groupMain, groupName sql.NullString
		var price, priceMember sql.NullFloat64

		if err := rows.Scan(&code, &barcode, &unitCode, &unitName, &itemType, &drinkType, &taxType, &havePoint, &icName, &groupMain, &groupName, &price, &priceMember); err != nil {
			log.Printf("Error scanning row for table %s: %v", tableName, err)
			continue
		}

		inventories = append(inventories, models.MongoProductBarcodeModel{
			Barcode:      barcode.String,
			ItemUnitCode: unitCode.String,
			ItemUnitNames: []models.LanguageNameModel{
				{Code: "th", Name: unitName.String},
			},
			ItemType:   int(itemType.Int64),
			FoodType:   int(drinkType.Int64),
			TaxType:    int(taxType.Int64),
			IsSumPoint: havePoint.Bool,
			ItemCode:   code.String,
			Names: []models.LanguageNameModel{
				{Code: "th", Name: icName.String},
			},
			GroupCode: groupMain.String,
			GroupNames: []models.LanguageNameModel{
				{Code: "th", Name: groupName.String},
			},
			Prices: []models.PriceModel{
				{KeyNumber: 1, Price: price.Float64},
				{KeyNumber: 2, Price: priceMember.Float64},
			},
		})
	}
	if err := rows.Err(); err != nil {
		logging.LogError(fmt.Sprintf("Error iterating rows for table %s", tableName), err)
		return err
	}
	log.Printf("Step 3: Processed %d items (%.2f seconds)", len(inventories), time.Since(stepStart).Seconds())

	// Step 4: Send data to API
	stepStart = time.Now()
	batchSize := 50
	totalItems := len(inventories)
	for i := 0; i < totalItems; i += batchSize {
		end := i + batchSize
		if end > totalItems {
			end = totalItems
		}
		batchStart := time.Now()

		responseBody, err := utils.SendDataToAPI("productbarcode", apiKey, inventories[i:end])
		if err != nil {
			logging.LogError(fmt.Sprintf("Error sending batch %d-%d for table %s", i+1, end, tableName), err)
			return err
		}

		var apiResponse struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
			logging.LogError(fmt.Sprintf("Error parsing API response for table %s (batch %d-%d)", tableName, i+1, end), err)
			continue
		}
		if !apiResponse.Success {
			logging.LogError(fmt.Sprintf("API reported failure for table %s (batch %d-%d)", tableName, i+1, end), fmt.Errorf(apiResponse.Message))
			continue
		}

		log.Printf("Sent batch %d-%d of %d for table %s (%.2f seconds)", i+1, end, totalItems, tableName, time.Since(batchStart).Seconds())
	}
	log.Printf("Step 4: Sent all data to API (%.2f seconds)", time.Since(stepStart).Seconds())

	duration := time.Since(start)
	logging.LogSuccess(fmt.Sprintf("Successful Table %s", tableName), databases.DatabaseName, duration, totalItems)
	logging.LogResult(databases.DatabaseName, fmt.Sprintf("Sync %s ToMongoDB", tableName), duration, totalItems)

	return nil
}
