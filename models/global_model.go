package models

import "time"

type DatabaseModel struct {
	DatabaseName string
}

type SyncResult struct {
	DatabaseName string
	FunctionName string
	Duration     time.Duration
	ItemCount    int
}
