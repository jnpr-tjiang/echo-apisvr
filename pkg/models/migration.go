package models

import (
	"gorm.io/gorm"
)

// MigrateDataModel - Auto migrate project models
func MigrateDataModel(db *gorm.DB) {
	if err := db.SetupJoinTable(&Device{}, "Devicefamilies", &DeviceDevicefamily{}); err != nil {
		panic(err)
	}
	db.AutoMigrate(&Domain{})
	db.AutoMigrate(&Project{})
	db.AutoMigrate(&Device{})
	db.AutoMigrate(&Devicefamily{})
}
