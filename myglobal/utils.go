package myglobal

import (
	"fmt"
	"log"
	"smlsynctodede/config"
	"smlsynctodede/models"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var PartServices = []models.PartService{
	{
		ServiceName: "creditor",
		PartName:    "debtaccount/creditor/bulk",
	},
	{
		ServiceName: "debtor",
		PartName:    "debtaccount/debtor/bulk",
	},
	{
		ServiceName: "productbarcode",
		PartName:    "product/barcode/bulk",
	},
}

var results []models.SyncResult
var resultMutex sync.Mutex

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
)

// GetPostgreSQLConnectionString returns the connection string for PostgreSQL
func GetPostgreSQLConnectionString(dbName string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.AppConfig.Database.Host,
		config.AppConfig.Database.Port,
		config.AppConfig.Database.User,
		config.AppConfig.Database.Password,
		dbName)
}

// GetFullAPIURL returns the full API URL for a given service
func GetFullAPIURL(serviceName string) string {
	for _, part := range PartServices {
		if part.ServiceName == serviceName {
			return config.AppConfig.API.BaseURL + part.PartName
		}
	}
	return ""
}

// GenerateGUID สร้าง GUID ใหม่
func GenerateGUID() string {
	return uuid.New().String()
}

// InitResults initializes the results slice
func InitResults() {
	results = make([]models.SyncResult, 0)
}

// LogResult บันทึกผลลัพธ์การทำงานของฟังก์ชัน
func LogResult(dbName, funcName string, duration time.Duration, count int) {
	resultMutex.Lock()
	defer resultMutex.Unlock()
	results = append(results, models.SyncResult{
		DatabaseName: dbName,
		FunctionName: funcName,
		Duration:     duration, // No rounding here
		ItemCount:    count,
	})
}

// PrintSummary prints the summary of the synchronization process
func PrintSummary(timeStart, timeStop time.Time) {
	log.Printf("%s=== Summary ===%s", ColorPurple, ColorReset)

	// สร้าง map เพื่อจัดกลุ่มผลลัพธ์ตาม database
	summaryByDB := make(map[string]map[string]models.SyncResult)
	var totalItems int

	// จัดกลุ่มผลลัพธ์ตาม database และฟังก์ชัน
	for _, result := range results {
		if _, exists := summaryByDB[result.DatabaseName]; !exists {
			summaryByDB[result.DatabaseName] = make(map[string]models.SyncResult)
		}
		summaryByDB[result.DatabaseName][result.FunctionName] = result
		totalItems += result.ItemCount
	}

	// แสดงผลสรุปแยกตาม database
	for dbName, dbResults := range summaryByDB {
		log.Printf("%sDatabase: %s%s", ColorBlue, dbName, ColorReset)
		var dbTotalItems int
		var dbTotalDuration time.Duration

		// สร้าง slice ของ keys เพื่อเรียงลำดับ
		var keys []string
		for k := range dbResults {
			keys = append(keys, k)
		}

		// เรียงลำดับตาม table name
		sort.Slice(keys, func(i, j int) bool {
			return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
		})

		for _, key := range keys {
			result := dbResults[key]
			log.Printf("  %-30s: %8d items, %s",
				result.FunctionName, result.ItemCount, FormatDuration(result.Duration))
			dbTotalItems += result.ItemCount
			dbTotalDuration += result.Duration
		}
		// แสดงผลรวมของแต่ละ database
		log.Printf("%s  %-30s: %8d items, %s%s\n",
			ColorYellow, "DB Total", dbTotalItems, FormatDuration(dbTotalDuration), ColorReset)
	}

	// แสดงผลรวมทั้งหมดและข้อมูลเวลาการทำงาน
	totalDuration := timeStop.Sub(timeStart)
	log.Printf("%sOverall Total      : %8d items%s", ColorRed, totalItems, ColorReset)
	log.Printf("Start Time         : %s%s%s", ColorCyan, timeStart.Format("2006-01-02 15:04:05.000"), ColorReset)
	log.Printf("End Time           : %s%s%s", ColorCyan, timeStop.Format("2006-01-02 15:04:05.000"), ColorReset)
	log.Printf("%sTotal Process Time : %s%s", ColorRed, FormatDuration(totalDuration), ColorReset)
}

// FormatDuration formats a duration to a human-readable string
func FormatDuration(d time.Duration) string {
	ms := d.Milliseconds() % 1000
	s := int(d.Seconds()) % 60
	m := int(d.Minutes()) % 60
	h := d.Hours()

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds %dms", int(h), int(m), int(s), int(ms))
	} else if m > 0 {
		return fmt.Sprintf("%dm %ds %dms", int(m), int(s), int(ms))
	}
	return fmt.Sprintf("%ds %dms", int(s), int(ms))
}

func LogError(message string, err error) {
	log.Printf("%s✗ %s: %v%s", ColorRed, message, err, ColorReset)
}

func LogSuccess(operation, databaseName string, duration time.Duration, itemCount int) {
	log.Printf("%s✓ %s: %s (%s, %d items)%s",
		ColorGreen, operation, databaseName, FormatDuration(duration), itemCount, ColorReset)
}

func GetIntInput(prompt string) int {
	for {
		input := getUserInput(prompt)
		num, err := strconv.Atoi(input)
		if err == nil {
			return num
		}
		fmt.Println("Invalid input. Please enter a valid number.")
	}
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return input
}
