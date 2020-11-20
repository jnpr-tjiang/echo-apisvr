package models

import (
	"gorm.io/gorm"
)

// MigrateDataModel - Auto migrate project models
func MigrateDataModel(db *gorm.DB) {
	if err := db.SetupJoinTable(&Device{}, "Device_families", &DeviceDevice_family{}); err != nil {
		panic(err)
	}
	db.AutoMigrate(&Domain{})
	db.AutoMigrate(&Project{})
	db.AutoMigrate(&Device{})
	db.AutoMigrate(&Device_family{})
}
