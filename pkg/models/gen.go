package models

import (
	"gorm.io/gorm"
)

// Init the data model metadata
func Init() {
	addHierarchy(Domain{}, Project{})
	addHierarchy(Project{}, Device{})
	addHierarchy(Project{}, Devicefamily{})
}

// Domain -
type Domain struct {
	Base BaseModel `gorm:"embedded"`
	// Has-Many relations
	Projects []Project `gorm:"foreignKey:ParentID;references:ID"`
}

// BeforeCreate to validate and set default
func (obj *Domain) BeforeCreate(tx *gorm.DB) error {
	return obj.Base.preCreate(tx, obj)
}

// Project -
type Project struct {
	Base BaseModel `gorm:"embedded"`
	// Has-Many relations
	Devices        []Device       `gorm:"foreignKey:ParentID;references:ID"`
	DeviceFamilies []Devicefamily `gorm:"foreignKey:ParentID;references:ID"`
}

// BeforeCreate to validate and set default
func (obj *Project) BeforeCreate(tx *gorm.DB) error {
	return obj.Base.preCreate(tx, obj)
}

// Devicefamily -
type Devicefamily struct {
	Base BaseModel `gorm:"embedded"`
	// Many-Many relations
	Devices []Device `gorm:"many2many:devicefamily_devices"`
}

// BeforeCreate to validate and set default
func (obj *Devicefamily) BeforeCreate(tx *gorm.DB) error {
	return obj.Base.preCreate(tx, obj)
}

// Device -
type Device struct {
	Base BaseModel `gorm:"embedded"`
	// Many-Many relations
	Devicefamilies []Devicefamily `gorm:"many2many:devicefamily_devices"`
}

// BeforeCreate to validate and set default
func (obj *Device) BeforeCreate(tx *gorm.DB) error {
	return obj.Base.preCreate(tx, obj)
}
