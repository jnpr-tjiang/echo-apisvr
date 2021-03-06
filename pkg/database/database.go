package database

import (
	"fmt"
	"time"

	"github.com/jnpr-tjiang/echo-apisvr/pkg/config"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db  *gorm.DB
	err error
)

// Init - initialize the database connection
func Init(cfg *config.Configuration) (*gorm.DB, error) {
	driver := cfg.Database.Driver
	dbName := cfg.Database.Dbname
	username := cfg.Database.Username
	host := cfg.Database.Host
	port := cfg.Database.Port

	if driver == "sqlite3" { // SQLITE
		dbFile := fmt.Sprintf("./%s.db", dbName)
		db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
			// Logger: logger.Default.LogMode(logger.Info),
		})
	} else if driver == "postgres" { // POSTGRES
		dsn := fmt.Sprintf("user=%s host=%s port=%s dbname=%s", username, host, port, dbName)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else { // MYSQL
		err = fmt.Errorf("%s is not supported yet", driver)
	}
	if err != nil {
		return nil, err
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDb.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDb.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDb.SetConnMaxLifetime(time.Duration(cfg.Database.MaxLifetime) * time.Second)

	models.MigrateDataModel(db)

	return db, nil
}

// GormDB - get gorm.DB
func GormDB() *gorm.DB {
	return db
}
