package myclickhouse

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func getClickHouseType(t reflect.Type) (string, bool) {
	switch t.Kind() {
	case reflect.String:
		return "String", true
	case reflect.Int:
		return "Int", true
	case reflect.Int8, reflect.Uint8:
		return "Int8", true
	case reflect.Int16, reflect.Int32, reflect.Uint16, reflect.Uint32:
		return "Int32", true
	case reflect.Int64, reflect.Uint64:
		return "Int64", true
	case reflect.Float32:
		return "Float32", true
	case reflect.Float64:
		return "Float64", true
	case reflect.TypeOf(time.Time{}).Kind():
		return "DateTime", true
	default:
		fmt.Println("Not support type", t.Kind())
		return "", false
	}
}

func QuerySelectAll(conn clickhouse.Conn, query string) ([]map[string]interface{}, error) {
	// ทำการ query ข้อมูล
	ctx := context.Background()
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	dataRows := make([]map[string]interface{}, 0)

	columnTypes := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	vars := make([]interface{}, len(columnTypes))
	for i := range columnTypes {
		vars[i] = reflect.New(columnTypes[i].ScanType()).Interface()
	}

	for rows.Next() {
		if err := rows.Scan(vars...); err != nil {
			return nil, err
		}

		rowData := make(map[string]interface{}, len(columnTypes))
		for i, v := range vars {
			rowData[columnTypes[i].Name()] = reflect.Indirect(reflect.ValueOf(v)).Interface()
		}

		dataRows = append(dataRows, rowData)
	}

	return dataRows, nil

	// 	// ดึงข้อมูลคอลัมน์ทั้งหมด
	// 	columns := rows.Columns()
	// 	columnCount := len(columns)

	// 	// สร้าง slice สำหรับเก็บผลลัพธ์ของแต่ละคอลัมน์
	// 	var results []map[string]interface{}

	// 	for rows.Next() {
	// 		values := make([]interface{}, columnCount)
	// 		valuePtrs := make([]interface{}, columnCount)
	// 		for i := range values {
	// 			valuePtrs[i] = &values[i]
	// 		}

	// 		if err := rows.Scan(valuePtrs...); err != nil {
	// 			return nil, fmt.Errorf("failed to scan row: %w", err)
	// 		}

	// 		rowMap := make(map[string]interface{})
	// 		for i, col := range columns {
	// 			switch v := values[i].(type) {
	// 			case *int64:
	// 				rowMap[col] = *v
	// 			case *float64:
	// 				rowMap[col] = *v
	// 			case *string:
	// 				rowMap[col] = *v
	// 			case *bool:
	// 				rowMap[col] = *v
	// 			default:
	// 				rowMap[col] = v
	// 			}
	// 		}

	// 		results = append(results, rowMap)
	// 	}

	//		return results, nil
	//	}
}

func Connect() (clickhouse.Conn, error) {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{"143.198.203.119:19000"},
			Auth: clickhouse.Auth{
				Database: "dedebi",
				Username: "smlchdb",
				Password: "heiR5XpDMyn4",
			},
		})
	)

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	return conn, nil
}

func ExecuteCommand(ctx context.Context, conn clickhouse.Conn, query string) error {
	err := conn.Exec(ctx, query)
	return err
}

