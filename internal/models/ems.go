package models

import "gorm.io/gorm"

// Domain -
type Domain struct {
	Base BaseModel `gorm:"embedded"`
	// Has-Many relations
	Projects []Project `gorm:"foreignKey:ParentID;references:ID"`
}

// BeforeCreate to validate and set default
func (domain *Domain) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// Project -
type Project struct {
	Base BaseModel `gorm:"embedded"`
	// Has-Many relations
	Devices        []Device       `gorm:"foreignKey:ParentID;references:ID"`
	DeviceFamilies []Devicefamily `gorm:"foreignKey:ParentID;references:ID"`
}

// BeforeCreate to validate and set default
func (project *Project) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// Devicefamily -
type Devicefamily struct {
	Base BaseModel `gorm:"embedded"`
	// Many-Many relations
	Devices []Device `gorm:"many2many:devicefamily_devices"`
}

// Device -
type Device struct {
	Base BaseModel `gorm:"embedded"`
	// Many-Many relations
	Devicefamilies []Devicefamily `gorm:"many2many:devicefamily_devices"`
}
