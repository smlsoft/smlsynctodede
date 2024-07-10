package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"smlsynctobc/myclickhouse"
	"smlsynctobc/myglobal"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/lib/pq"
)

// customDecodeHookFunc ปรับให้รองรับการแปลงวันที่แบบ ClickHouse
func customDecodeHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() == reflect.Map {
		mapData := data.(map[string]interface{})
		if val, ok := mapData["$numberInt"]; ok && t.Kind() == reflect.Int {
			return int(val.(float64)), nil
		}
		if val, ok := mapData["$numberDouble"]; ok && t.Kind() == reflect.Float64 {
			switch val.(type) {
			case string:
				return fmt.Sscanf(val.(string), "%f")
			default:
				return val.(float64), nil
			}
		}
		if val, ok := mapData["$numberLong"]; ok && t == reflect.TypeOf(time.Time{}) {
			ms, _ := time.ParseDuration(fmt.Sprintf("%sms", val.(string)))
			return time.Unix(0, ms.Nanoseconds()), nil
		}
		if val, ok := mapData["$date"]; ok && t == reflect.TypeOf(time.Time{}) {
			switch val.(type) {
			case string:
				return time.Parse(time.RFC3339, val.(string))
			case map[string]interface{}:
				if dateVal, ok := val.(map[string]interface{})["$numberLong"]; ok {
					switch dateVal.(type) {
					case string:
						ms, _ := time.ParseDuration(fmt.Sprintf("%sms", dateVal.(string)))
						return time.Unix(0, ms.Nanoseconds()), nil
					}
				}
			}
		}
	}

	if f.Kind() == reflect.String && t == reflect.TypeOf(time.Time{}) {
		return time.Parse(time.RFC3339, data.(string))
	}

	return data, nil
}

func productBarcodeRebuildInsertToClickHouse(clickHouseConn clickhouse.Conn, clickHouseData []string) {
	queryInsert := "INSERT INTO dedebi.productbarcode (shopid, barcode, itemcode, name0) VALUES " + strings.Join(clickHouseData, ",")
	err := myclickhouse.ExecuteCommand(context.Background(), clickHouseConn, queryInsert)
	if err != nil {
		log.Println("Error insert into ClickHouse : " + queryInsert)
		log.Fatal(err)
	}
}

func productBarcodeRebuild(database myglobal.DatabaseModel) {
	timeStart := time.Now()
	fmt.Println(database.DatabaseName + " : productBarcodeRebuild Start " + timeStart.String())

	// ทำการเชื่อมต่อ PostgreSQL
	connPostgreSqlStr := myglobal.GetPostgreSQLConnectionString(database.DatabaseName)

	// เปิดการเชื่อมต่อ
	db, err := sql.Open("postgres", connPostgreSqlStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ทดสอบการเชื่อมต่อ
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database!")

	// ลบข้อมูลเก่า ออกจาก ClickHouse
	clickHouseConn, _ := myclickhouse.Connect()
	err = myclickhouse.ExecuteCommand(context.Background(), clickHouseConn, "alter table dedebi.productbarcode delete where shopid = '"+database.ShopId+"'")
	if err != nil {
		log.Println("Error truncate ClickHouse")
		log.Fatal(err)
	}
	clickHouseData := []string{}

	// ตัวอย่างการ query
	rows, err := db.Query("SELECT ic_code,barcode,description,unit_code FROM ic_inventory_barcode")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// ประมวลผลข้อมูลที่ได้จาก query
	count := 0
	for rows.Next() {
		var ic_code string
		var barcode string
		var description string
		var unit_code string
		err = rows.Scan(&ic_code, &barcode, &description, &unit_code)
		if err != nil {
			log.Fatal(err)
		}
		productName := "Error"
		if len(description) > 0 {
			productName = description
		}
		productName = strings.Replace(productName, "'", "''", -1)
		clickHouseData = append(clickHouseData, "('"+database.ShopId+"', '"+barcode+"', '"+ic_code+"', '"+productName+"')")
		count++
		if count%10000 == 0 {
			productBarcodeRebuildInsertToClickHouse(clickHouseConn, clickHouseData)
			clickHouseData = []string{}
		}
	}
	if len(clickHouseData) > 0 {
		productBarcodeRebuildInsertToClickHouse(clickHouseConn, clickHouseData)
	}

	clickHouseConn.Close()
	timeStop := time.Now()
	timeDiffStr := timeStop.Sub(timeStart).String()
	fmt.Println(database.DatabaseName + " : productBarcodeRebuild done : " + timeDiffStr)
}

// หัวรายวัน

func transDocRebuildInsert(clickHouseConn clickhouse.Conn, clickHouseDoc *[]string) {
	queryInsert := "INSERT INTO dedebi.doc (shopid,branchid, docno, docdatetime,perioddatetime, totalamount,paycashamount,paycashchange,paycashbalance,roundamount) VALUES " + strings.Join(*clickHouseDoc, ",")
	err := myclickhouse.ExecuteCommand(context.Background(), clickHouseConn, queryInsert)
	if err != nil {
		log.Println("Error insert into ClickHouse : " + queryInsert)
		log.Fatal(err)
	}
}

func transDocRebuild(database myglobal.DatabaseModel) {
	timeStart := time.Now()
	fmt.Println(database.DatabaseName + " : transSaleInvoiceRebuild Start" + timeStart.String())
	clickHouseConn, _ := myclickhouse.Connect()
	tableList := []string{"dedebi.doc"}
	for _, table := range tableList {
		query := fmt.Sprintf("ALTER TABLE %s DELETE WHERE shopid = '%s'", table, database.ShopId)
		err := myclickhouse.ExecuteCommand(context.Background(), clickHouseConn, query)
		if err != nil {
			log.Printf("Error truncating ClickHouse table %s: %v", table, err)
			log.Fatal(err)
		}
	}

	// ทำการเชื่อมต่อ PostgreSQL
	connPostgreSqlStr := myglobal.GetPostgreSQLConnectionString(database.DatabaseName)

	// เปิดการเชื่อมต่อ
	db, err := sql.Open("postgres", connPostgreSqlStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ทดสอบการเชื่อมต่อ
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database!")

	// ตัวอย่างการ query
	icTransrows, err := db.Query("SELECT doc_date,doc_no,total_amount FROM ic_trans")
	if err != nil {
		log.Fatal(err)
	}
	defer icTransrows.Close()

	// ประมวลผลข้อมูลที่ได้จาก query
	count := 0
	clickHouseDoc := []string{}
	for icTransrows.Next() {
		var docDateStr string
		var docNo string
		var totalAmount float64
		var payCashAmount float64
		var payCashChange float64
		var roundAmount float64

		err = icTransrows.Scan(&docDateStr, &docNo, &totalAmount)
		if err != nil {
			log.Fatal(err)
		}
		payCashAmount = 0
		payCashChange = 0
		roundAmount = 0

		// แปลง docDateStr เป็น time.Time โดยใช้ layout ที่ถูกต้อง
		docDate, err := time.Parse(time.RFC3339, docDateStr)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			continue
		}

		// ฟอร์แมตวันที่ให้เป็นรูปแบบที่ ClickHouse ยอมรับ
		docDateTimeStr := docDate.Format("2006-01-02 15:04:05")
		branchId := ""
		// เอกสาร
		clickHouseDoc = append(clickHouseDoc, fmt.Sprintf("('%s','%s','%s','%s','%s',%.2f,%.2f,%.2f,%.2f,%.2f)", database.ShopId, branchId, docNo, docDateTimeStr, docDateTimeStr, totalAmount, payCashAmount, payCashChange, 0.0, roundAmount))
		// แบบเก็บข้อมูลทุก 10000 รายการ
		count++
		if count%10000 == 0 {
			fmt.Println("doc count : " + fmt.Sprintf("%d", count))
			transDocRebuildInsert(clickHouseConn, &clickHouseDoc)
			clickHouseDoc = []string{}
		}
	}
	if len(clickHouseDoc) > 0 {
		transDocRebuildInsert(clickHouseConn, &clickHouseDoc)
	}
	clickHouseConn.Close()
	timeStop := time.Now()
	timeDiffStr := timeStop.Sub(timeStart).String()
	fmt.Println("transSaleInvoiceRebuild done : " + timeDiffStr + " doc count : " + fmt.Sprintf("%d", count))
}

// รายวันย่อย

func transDocDetailRebuildInsert(clickHouseConn clickhouse.Conn, clickHouseDocDetail *[]string) {
	queryInsertDocDetail := "INSERT INTO dedebi.docdetail (shopid, branchid,docno, docdatetime, perioddatetime,barcode,  qty, price, sumamount, discountamount) VALUES " + strings.Join(*clickHouseDocDetail, ",")
	err := myclickhouse.ExecuteCommand(context.Background(), clickHouseConn, queryInsertDocDetail)
	if err != nil {
		log.Println("Error insert into ClickHouse : " + queryInsertDocDetail)
		log.Fatal(err)
	}
}

func transDocDetailRebuild(database myglobal.DatabaseModel) {
	timeStart := time.Now()
	fmt.Println(database.DatabaseName + " : transDocDetailRebuild Start" + timeStart.String())
	clickHouseConn, _ := myclickhouse.Connect()
	tableList := []string{"dedebi.docdetail"}
	for _, table := range tableList {
		query := fmt.Sprintf("ALTER TABLE %s DELETE WHERE shopid = '%s'", table, database.ShopId)
		err := myclickhouse.ExecuteCommand(context.Background(), clickHouseConn, query)
		if err != nil {
			log.Printf("Error truncating ClickHouse table %s: %v", table, err)
			log.Fatal(err)
		}
	}

	// ทำการเชื่อมต่อ PostgreSQL
	connPostgreSqlStr := myglobal.GetPostgreSQLConnectionString(database.DatabaseName)

	// เปิดการเชื่อมต่อ
	db, err := sql.Open("postgres", connPostgreSqlStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ทดสอบการเชื่อมต่อ
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database!")

	// ตัวอย่างการ query หลายบรรทัด
	query := `
			
			SELECT 
			doc_date,
			doc_no,
			COALESCE(COALESCE(NULLIF(barcode, ''), item_code),'') AS barcode,
			qty,
			price,
			sum_amount,
			discount_amount 
		FROM 
			ic_trans_detail 
			
			`

	icTransrows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer icTransrows.Close()

	// ประมวลผลข้อมูลที่ได้จาก query
	count := 0
	clickHouseDocDetail := []string{}
	for icTransrows.Next() {
		var docDateStr string
		var docNo string
		var barcode string
		var qty sql.NullFloat64
		var price sql.NullFloat64
		var sumAmount sql.NullFloat64
		var discountAmount sql.NullFloat64

		err = icTransrows.Scan(&docDateStr, &docNo, &barcode, &qty, &price, &sumAmount, &discountAmount)
		if err != nil {
			log.Fatal(err)
		}

		// แปลง docDateStr เป็น time.Time โดยใช้ layout ที่ถูกต้อง
		docDate, err := time.Parse(time.RFC3339, docDateStr)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			continue
		}

		// ฟอร์แมตวันที่ให้เป็นรูปแบบที่ ClickHouse ยอมรับ
		docDateTimeStr := docDate.Format("2006-01-02 15:04:05")
		branchId := ""
		// เอกสาร
		clickHouseDocDetail = append(clickHouseDocDetail, fmt.Sprintf("('%s','%s','%s','%s','%s','%s',%.2f,%.2f,%.2f,%.2f)", database.ShopId, branchId, docNo, docDateTimeStr, docDateTimeStr, barcode, myglobal.NullFloat64ToFloat(qty), myglobal.NullFloat64ToFloat(price), myglobal.NullFloat64ToFloat(sumAmount), myglobal.NullFloat64ToFloat(discountAmount)))
		// แบบเก็บข้อมูลทุก 10000 รายการ
		count++
		if count%100000 == 0 {
			fmt.Println("doc count : " + fmt.Sprintf("%d", count))
			transDocDetailRebuildInsert(clickHouseConn, &clickHouseDocDetail)
			clickHouseDocDetail = []string{}
		}
	}
	if len(clickHouseDocDetail) > 0 {
		transDocDetailRebuildInsert(clickHouseConn, &clickHouseDocDetail)
	}
	clickHouseConn.Close()
	timeStop := time.Now()
	timeDiffStr := timeStop.Sub(timeStart).String()
	fmt.Println("transDocDetailRebuild done : " + timeDiffStr + " doc count : " + fmt.Sprintf("%d", count))
}

/// ======= sml  ======

// / รายวันย่อย ic_trans_detail
// รายวันย่อย

func icTransDetailRebuildInsertToClikeHouse(clickHouseConn clickhouse.Conn, clickHouseDocDetail *[]string) {
	queryInsertDocDetail := `
	INSERT INTO dedebi.ic_trans_detail (
		doc_date, doc_time, doc_no, trans_flag, item_code, 
		unit_code, wh_code, shelf_code, calc_flag, inquiry_type, 
		doc_ref, is_pos, stand_value, divide_value, qty, 
		sum_of_cost, profit_lost_cost_amount, last_status, item_type, is_doc_copy, 
		sum_amount, price, discount, ref_doc_no, item_code_main, 
		ref_guid, set_ref_line , shopid
	) VALUES ` + strings.Join(*clickHouseDocDetail, ",")

	err := myclickhouse.ExecuteCommand(context.Background(), clickHouseConn, queryInsertDocDetail)
	if err != nil {
		log.Println("Error insert into ClickHouse : " + queryInsertDocDetail)
		log.Fatal(err)
	}
}

func icTransDetailRebuild(database myglobal.DatabaseModel) {
	timeStart := time.Now()
	fmt.Println(database.DatabaseName + " : icTransDetailRebuild Start" + timeStart.String())
	clickHouseConn, _ := myclickhouse.Connect()
	tableList := []string{"dedebi.ic_trans_detail"}
	for _, table := range tableList {
		query := fmt.Sprintf("ALTER TABLE %s DELETE WHERE shopid = '%s'", table, database.ShopId)
		err := myclickhouse.ExecuteCommand(context.Background(), clickHouseConn, query)
		if err != nil {
			log.Printf("Error truncating ClickHouse table %s: %v", table, err)
			log.Fatal(err)
		}
	}

	// ทำการเชื่อมต่อ PostgreSQL
	connPostgreSqlStr := myglobal.GetPostgreSQLConnectionString(database.DatabaseName)

	// เปิดการเชื่อมต่อ
	db, err := sql.Open("postgres", connPostgreSqlStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ทดสอบการเชื่อมต่อ
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database!")

	// ดึงข้อมูลจาก PostgreSQL
	query := `
	SELECT 
		doc_date, doc_time, doc_no, trans_flag, item_code, 
		unit_code, wh_code, shelf_code, calc_flag, inquiry_type, 
		doc_ref, is_pos, stand_value, divide_value, qty, 
		sum_of_cost, profit_lost_cost_amount, last_status, item_type, is_doc_copy, 
		sum_amount, price, discount, ref_doc_no, item_code_main, 
		ref_guid, set_ref_line
	FROM 
		ic_trans_detail
	`

	icTransrows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer icTransrows.Close()

	// ประมวลผลข้อมูลที่ได้จาก query
	count := 0
	clickHouseDocDetail := []string{}
	for icTransrows.Next() {
		var docDate, docTime, docNo, itemCode, unitCode, whCode, shelfCode, docRef, discount, refDocNo, itemCodeMain, refGuid, setRefLine sql.NullString
		var transFlag, calcFlag, inquiryType, isPos, lastStatus, itemType, isDocCopy sql.NullInt32
		var standValue, divideValue, qty, sumOfCost, profitLostCostAmount, sumAmount, price sql.NullFloat64

		err = icTransrows.Scan(
			&docDate, &docTime, &docNo, &transFlag, &itemCode,
			&unitCode, &whCode, &shelfCode, &calcFlag, &inquiryType,
			&docRef, &isPos, &standValue, &divideValue, &qty,
			&sumOfCost, &profitLostCostAmount, &lastStatus, &itemType, &isDocCopy,
			&sumAmount, &price, &discount, &refDocNo, &itemCodeMain,
			&refGuid, &setRefLine,
		)
		if err != nil {
			log.Fatal(err)
		}

		// แปลงวันที่และเวลาให้เป็นรูปแบบที่ ClickHouse ยอมรับ
		var docDateStr, docTimeStr string
		if docDate.Valid {
			// แปลงวันที่จากฟอร์แมต ISO 8601 เป็นฟอร์แมตที่ต้องการ
			t, err := time.Parse(time.RFC3339[:10], docDate.String[:10])
			if err != nil {
				log.Printf("Error parsing date: %v", err)
				continue
			}
			docDateStr = t.Format("2006-01-02")
		} else {
			docDateStr = "0000-00-00"
		}

		if docTime.Valid {
			// ใช้เวลาจาก docTime ถ้ามี
			docTimeStr = docTime.String
		} else {
			// ถ้าไม่มี docTime ให้ใช้เวลาจาก docDate ถ้ามี
			if docDate.Valid {
				t, err := time.Parse(time.RFC3339, docDate.String)
				if err != nil {
					log.Printf("Error parsing time from date: %v", err)
					docTimeStr = "00:00:00"
				} else {
					docTimeStr = t.Format("15:04:05")
				}
			} else {
				docTimeStr = "00:00:00"
			}
		}
		// เอกสาร
		// สร้าง string สำหรับ insert โดยจัดการกับค่า NULL
		clickHouseDocDetail = append(clickHouseDocDetail, fmt.Sprintf("('%s','%s','%s',%d,'%s','%s','%s','%s',%d,%d,'%s',%d,%.2f,%.2f,%.2f,%.2f,%.2f,%d,%d,%d,%.2f,%.2f,'%s','%s','%s','%s','%s','%s')",
			docDateStr,
			docTimeStr,
			myglobal.NullStringToString(docNo),
			myglobal.NullInt32ToInt(transFlag),
			myglobal.NullStringToString(itemCode),
			myglobal.NullStringToString(unitCode),
			myglobal.NullStringToString(whCode),
			myglobal.NullStringToString(shelfCode),
			myglobal.NullInt32ToInt(calcFlag),
			myglobal.NullInt32ToInt(inquiryType),
			myglobal.NullStringToString(docRef),
			myglobal.NullInt32ToInt(isPos),
			myglobal.NullFloat64ToFloat(standValue),
			myglobal.NullFloat64ToFloat(divideValue),
			myglobal.NullFloat64ToFloat(qty),
			myglobal.NullFloat64ToFloat(sumOfCost),
			myglobal.NullFloat64ToFloat(profitLostCostAmount),
			myglobal.NullInt32ToInt(lastStatus),
			myglobal.NullInt32ToInt(itemType),
			myglobal.NullInt32ToInt(isDocCopy),
			myglobal.NullFloat64ToFloat(sumAmount),
			myglobal.NullFloat64ToFloat(price),
			myglobal.NullStringToString(discount),
			myglobal.NullStringToString(refDocNo),
			myglobal.NullStringToString(itemCodeMain),
			myglobal.NullStringToString(refGuid),
			myglobal.NullStringToString(setRefLine),
			database.ShopId))

		// แบบเก็บข้อมูลทุก 10000 รายการ
		count++
		if count%100000 == 0 {
			fmt.Println("doc count : " + fmt.Sprintf("%d", count))
			icTransDetailRebuildInsertToClikeHouse(clickHouseConn, &clickHouseDocDetail)
			clickHouseDocDetail = []string{}
		}
	}
	if len(clickHouseDocDetail) > 0 {
		icTransDetailRebuildInsertToClikeHouse(clickHouseConn, &clickHouseDocDetail)
	}
	clickHouseConn.Close()
	timeStop := time.Now()
	timeDiffStr := timeStop.Sub(timeStart).String()
	fmt.Println("icTransDetailRebuild done : " + timeDiffStr + " doc count : " + fmt.Sprintf("%d", count))
}

// / ข้อมูลสินค้า ic_inventory
// icInventoryRebuildInsertToClickHouse เพิ่มข้อมูลสินค้าลงใน ClickHouse
func icInventoryRebuildInsertToClickHouse(clickHouseConn clickhouse.Conn, clickHouseData []string) {
	queryInsert := "INSERT INTO dedebi.ic_inventory (code, name_1, unit_cost, unit_standard ,item_type , shopid) VALUES " + strings.Join(clickHouseData, ",")
	err := myclickhouse.ExecuteCommand(context.Background(), clickHouseConn, queryInsert)
	if err != nil {
		log.Println("Error insert into ClickHouse : " + queryInsert)
		log.Fatal(err)
	}
}

func icInventoryRebuild(database myglobal.DatabaseModel) {
	timeStart := time.Now()
	fmt.Println(database.DatabaseName + " : icInventoryRebuild Start " + timeStart.String())

	// ทำการเชื่อมต่อ PostgreSQL
	connPostgreSqlStr := myglobal.GetPostgreSQLConnectionString(database.DatabaseName)

	// เปิดการเชื่อมต่อ
	db, err := sql.Open("postgres", connPostgreSqlStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ทดสอบการเชื่อมต่อ
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database!")

	// ลบข้อมูลเก่า ออกจาก ClickHouse
	clickHouseConn, _ := myclickhouse.Connect()
	err = myclickhouse.ExecuteCommand(context.Background(), clickHouseConn, "alter table dedebi.ic_inventory delete where shopid = '"+database.ShopId+"'")
	if err != nil {
		log.Println("Error truncate ClickHouse")
		log.Fatal(err)
	}
	clickHouseData := []string{}

	// ตัวอย่างการ query
	rows, err := db.Query("SELECT code, name_1, unit_cost, unit_standard, item_type FROM ic_inventory")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// ประมวลผลข้อมูลที่ได้จาก query
	count := 0
	for rows.Next() {
		var code string
		var name_1 string
		var unit_cost string
		var unit_standard string
		var item_type int
		err = rows.Scan(&code, &name_1, &unit_cost, &unit_standard, &item_type)
		if err != nil {
			log.Fatal(err)
		}
		// Replace single quotes with two single quotes for SQL escaping
		name_1 = strings.ReplaceAll(name_1, "'", "''")

		clickHouseData = append(clickHouseData, "( '"+code+"', '"+name_1+"', '"+unit_cost+"' , '"+unit_standard+"' , 0, '"+database.ShopId+"')")
		count++
		if count%10000 == 0 {
			icInventoryRebuildInsertToClickHouse(clickHouseConn, clickHouseData)
			clickHouseData = []string{}
		}
	}
	if len(clickHouseData) > 0 {
		icInventoryRebuildInsertToClickHouse(clickHouseConn, clickHouseData)
	}

	clickHouseConn.Close()
	timeStop := time.Now()
	timeDiffStr := timeStop.Sub(timeStart).String()
	fmt.Println("icInventoryRebuild done : " + timeDiffStr + " item count : " + fmt.Sprintf("%d", count))
}

func main() {
	timeStart := time.Now()
	log.Printf("Start Time: %s", timeStart)

	for _, database := range myglobal.DatabaseList {

		productBarcodeRebuild(database)

		transDocRebuild(database)

		transDocDetailRebuild(database)

		// === stock sml ===
		icInventoryRebuild(database)

		icTransDetailRebuild(database)

	}

	timeStop := time.Now()
	log.Printf("End Time: %s", timeStop)
	log.Printf("Total Time: %s", timeStop.Sub(timeStart))

}
