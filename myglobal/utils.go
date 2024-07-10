package myglobal

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
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
