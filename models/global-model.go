package models

import (
	"time"
)

type DatabaseModel struct {
	DatabaseName string
}

type PartService struct {
	ServiceName string
	PartName    string
}

type LanguageNameModel struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Isauto   bool   `json:"isauto"`
	Isdelete bool   `json:"isdelete"`
}

type PriceModel struct {
	KeyNumber int     `json:"keynumber"`
	Price     float64 `json:"price"`
}

// SyncResult เก็บผลลัพธ์การทำงานของแต่ละฟังก์ชัน
type SyncResult struct {
	DatabaseName string
	FunctionName string
	Duration     time.Duration
	ItemCount    int
}
