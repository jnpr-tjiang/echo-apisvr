package models

import (
	"gorm.io/gorm"
)

// Domain -
type Domain struct {
	Base BaseModel `gorm:"embedded"`
	// Has-Many relations
	Projects []Project `gorm:"foreignKey:ParentID;references:ID"`
}

// ParentType of the parent object
func (obj *Domain) ParentType() string {
	return ""
}

// BeforeCreate to validate and set default
func (obj *Domain) BeforeCreate(tx *gorm.DB) error {
	return obj.Base.preCreate(tx, obj.ParentType())
}

// Project -
type Project struct {
	Base BaseModel `gorm:"embedded"`
	// Has-Many relations
	Devices        []Device       `gorm:"foreignKey:ParentID;references:ID"`
	DeviceFamilies []Devicefamily `gorm:"foreignKey:ParentID;references:ID"`
}

// ParentType of the parent object
func (obj *Project) ParentType() string {
	return "Domain"
}

// BeforeCreate to validate and set default
func (obj *Project) BeforeCreate(tx *gorm.DB) error {
	return obj.Base.preCreate(tx, obj.ParentType())
}

// Devicefamily -
type Devicefamily struct {
	Base BaseModel `gorm:"embedded"`
	// Many-Many relations
	Devices []Device `gorm:"many2many:devicefamily_devices"`
}

// ParentType of the parent object
func (obj *Devicefamily) ParentType() string {
	return "Project"
}

// BeforeCreate to validate and set default
func (obj *Devicefamily) BeforeCreate(tx *gorm.DB) error {
	return obj.Base.preCreate(tx, obj.ParentType())
}

// Device -
type Device struct {
	Base BaseModel `gorm:"embedded"`
	// Many-Many relations
	Devicefamilies []Devicefamily `gorm:"many2many:devicefamily_devices"`
}

// ParentType of the parent object
func (obj *Device) ParentType() string {
	return "Project"
}

// BeforeCreate to validate and set default
func (obj *Device) BeforeCreate(tx *gorm.DB) error {
	return obj.Base.preCreate(tx, obj.ParentType())
}
