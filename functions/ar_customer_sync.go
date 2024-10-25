package functions

import (
	"database/sql"
	"fmt"
	"log"
	"smlsynctodede/database"
	"smlsynctodede/logging"
	"smlsynctodede/models"
	"smlsynctodede/utils"
	"strconv"
	"time"
)

func SyncArCustomerToMongoDB(databases models.DatabaseModel, apiKey string) error {
	tableName := "ar_customer"
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
	log.Printf("Step 1: Connected to PostgreSQL (%.2f seconds)", time.Since(stepStart).Seconds())

	// Step 2: Query data from PostgreSQL
	stepStart = time.Now()
	rows, err := db.Query(`
		SELECT 
		a.code
		,a.name_1
		,a.ar_status
		,b.tax_id
		,b.branch_type
		,b.branch_code
		,a.email
		,b.credit_day
		,case when (select count(ar_code) from ar_dealer where a.code = ar_dealer.ar_code) > 0 then 1 else 0 end as member_status 
		,a.address
		,a.telephone
		FROM ar_customer a,ar_customer_detail b
		WHERE a.code = b.ar_code
		ORDER BY a.code
		`)
	if err != nil {
		logging.LogError(fmt.Sprintf("Error querying PostgreSQL for table %s", tableName), err)
		return err
	}
	defer rows.Close()
	log.Printf("Step 2: Queried data from PostgreSQL (%.2f seconds)", time.Since(stepStart).Seconds())

	// Step 3: Process query results
	stepStart = time.Now()
	var debtors []models.MongoDebtorModel
	for rows.Next() {
		var code, name, arStatus, taxId, branchType, branchCode, email, creditDay, memberStatus, address, telephone sql.NullString
		err := rows.Scan(&code, &name, &arStatus, &taxId, &branchType, &branchCode, &email, &creditDay, &memberStatus, &address, &telephone)
		if err != nil {
			log.Printf("Error scanning row for table %s: %v", tableName, err)
			continue
		}

		personalType := 1 // Default to 1 (บุคคลธรรมดา)
		if arStatus.String == "1" {
			personalType = 2 // Set to 2 (นิติบุคคล) if ar_status is 1
		}

		customerType := 1 // Default to 1 (สำนักงานใหญ่)
		branchNumber := "00000"
		if branchType.String == "1" {
			customerType = 2 // Set to 2 (สาขา) if branch_type is 1
			branchNumber = branchCode.String
		}
		creditDayInt, _ := strconv.Atoi(creditDay.String)
		isMember := memberStatus.String == "1"

		debtor := models.MongoDebtorModel{
			Code:         code.String,
			Names:        []models.LanguageNameModel{{Code: "th", Name: name.String, Isauto: false, Isdelete: false}},
			PersonalType: personalType,
			TaxId:        taxId.String,
			CustomerType: customerType,
			BranchNumber: branchNumber,
			Email:        email.String,
			CreditDay:    creditDayInt,
			IsMember:     isMember,
			AddressForBilling: models.CustomerAddressModel{
				Contactnames: []models.LanguageNameModel{{Code: "th", Name: name.String, Isauto: false, Isdelete: false}},
				Address:      []string{address.String},
				Phoneprimary: telephone.String,
			},
			Groups: []string{},
			Images: []string{},
		}

		debtors = append(debtors, debtor)
	}
	log.Printf("Step 3: Processed %d items (%.2f seconds)", len(debtors), time.Since(stepStart).Seconds())

	// Step 4: Send data to API
	stepStart = time.Now()
	batchSize := 50
	totalItems := len(debtors)
	for i := 0; i < totalItems; i += batchSize {
		batchStart := time.Now()
		end := i + batchSize
		if end > totalItems {
			end = totalItems
		}
		batch := debtors[i:end]

		responseBody, err := utils.SendDataToAPI("debtor", apiKey, batch)
		if err != nil {
			if err.Error() == "API request failed with status code 401: {\"message\":\"Token Invalid.\",\"success\":false}" {
				logging.LogError("Authentication failed. Please check your API key.", fmt.Errorf(string(responseBody)))
				return err // Return error to stop the synchronization process
			}
			logging.LogError(fmt.Sprintf("Error sending data to API for table %s (batch %d-%d)", tableName, i+1, end), err)
			return err // Return error to stop the synchronization process
		}

		batchDuration := time.Since(batchStart)
		log.Printf("Sent batch %d-%d of %d for table %s (%.2f seconds)", i+1, end, totalItems, tableName, batchDuration.Seconds())
	}
	log.Printf("Step 4: Sent all data to API (%.2f seconds)", time.Since(stepStart).Seconds())

	duration := time.Since(start)
	logging.LogSuccess(fmt.Sprintf("Successful Table %s", tableName), databases.DatabaseName, duration, totalItems)
	logging.LogResult(databases.DatabaseName, fmt.Sprintf("Sync %s ToMongoDB", tableName), duration, totalItems)

	return nil
}
