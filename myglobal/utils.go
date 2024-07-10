package myglobal

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

var ShopId = "2QoilMQkX9i6vtAE88ilEubnrhz"

func CalculateSHA256(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

// PostgreSQL connection parameters
var (
	/// Jead DB :

	// PgHost     = "localhost"
	// PgPort     = 5432
	// PgUser     = "postgres"
	// PgPassword = "19682511"

	PgHost     = "192.168.2.236"
	PgPort     = 5432
	PgUser     = "postgres"
	PgPassword = "sml"
)

type DatabaseModel struct {
	DatabaseName string
	ShopId       string
}

var DatabaseList = []DatabaseModel{
	// {
	// 	DatabaseName: "data03",
	// 	ShopId:       "VDATA03",
	// },
	// {
	// 	DatabaseName: "data04",
	// 	ShopId:       "VDATA04",
	// },

	/// stock sml test
	{
		DatabaseName: "sml1_old",
		ShopId:       "sml1_old",
	},
}

// GetPostgreSQLConnectionString returns the connection string for PostgreSQL
func GetPostgreSQLConnectionString(dbName string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		PgHost, PgPort, PgUser, PgPassword, dbName)
}

// NullStringToString converts sql.NullString to string
func NullStringToString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

// NullInt32ToInt converts sql.NullInt32 to int
func NullInt32ToInt(i sql.NullInt32) int {
	if i.Valid {
		return int(i.Int32)
	}
	return 0
}

// NullFloat64ToFloat converts sql.NullFloat64 to float64
func NullFloat64ToFloat(f sql.NullFloat64) float64 {
	if f.Valid {
		return f.Float64
	}
	return 0.00
}

// FormatDateTime แปลงค่า docDate และ docTime จาก sql.NullString เป็นรูปแบบที่เหมาะสมสำหรับ ClickHouse
func FormatDateTime(docDate, docTime sql.NullString) (string, string) {
	var docDateStr, docTimeStr string

	// จัดการกับวันที่
	if docDate.Valid {
		// แปลงวันที่จากฟอร์แมต ISO 8601 เป็นฟอร์แมตที่ต้องการ
		t, err := time.Parse(time.RFC3339[:10], docDate.String[:10])
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			docDateStr = "0000-00-00"
		} else {
			docDateStr = t.Format("2006-01-02")
		}
	} else {
		docDateStr = "0000-00-00"
	}

	// จัดการกับเวลา
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

	return docDateStr, docTimeStr
}
