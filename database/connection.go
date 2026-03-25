package database

import (
	"database/sql"
	"fmt"
	"smlsynctodede/config"
)

func TestPostgresConnection(cfg config.Config) error {
	for _, db := range cfg.Databases {
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, db.Name)

		testDB, err := sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("error opening connection to %s: %v", db.Name, err)
		}

		err = testDB.Ping()
		testDB.Close()
		if err != nil {
			return fmt.Errorf("error connecting to %s: %v", db.Name, err)
		}
		fmt.Printf("Successfully connected to database: %s\n", db.Name)
	}
	return nil
}

func GetPostgreSQLConnectionString(dbName string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.AppConfig.Database.Host,
		config.AppConfig.Database.Port,
		config.AppConfig.Database.User,
		config.AppConfig.Database.Password,
		dbName)
}
