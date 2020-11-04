package models

import (
	"gorm.io/gorm"
)

// Init the data model's meta info
func Init() error {
	register("domain", modelInfo{
		allowedParentTypes: []string{},
		constructor:        func() Entity { return &Domain{} },
	})
	register("project", modelInfo{
		allowedParentTypes: []string{"domain"},
		constructor:        func() Entity { return &Project{} },
	})
	register("device", modelInfo{
		allowedParentTypes: []string{"project"},
		constructor:        func() Entity { return &Device{} },
	})
	register("devicefamily", modelInfo{
		allowedParentTypes: []string{"project"},
		constructor:        func() Entity { return &Devicefamily{} },
	})

	return nil
}

// Domain -----------------------------------------------------------------
type Domain struct {
	Base BaseModel `gorm:"embedded"`
	// Has-Many relations
	Projects []Project `gorm:"foreignKey:ParentID;references:ID"`
}

// BaseModel returns the reference to the base model
func (entity *Domain) BaseModel() *BaseModel {
	return &entity.Base
}

// BeforeCreate to validate and set default
func (entity *Domain) BeforeCreate(tx *gorm.DB) error {
	return entity.Base.preCreate(tx, entity)
}

// Project -----------------------------------------------------------------
type Project struct {
	Base BaseModel `gorm:"embedded"`
	// Has-Many relations
	Devices        []Device       `gorm:"foreignKey:ParentID;references:ID"`
	DeviceFamilies []Devicefamily `gorm:"foreignKey:ParentID;references:ID"`
}

// BaseModel returns the reference to the base model
func (entity *Project) BaseModel() *BaseModel {
	return &entity.Base
}

// BeforeCreate to validate and set default
func (entity *Project) BeforeCreate(tx *gorm.DB) error {
	return entity.Base.preCreate(tx, entity)
}

// Devicefamily -----------------------------------------------------------------
type Devicefamily struct {
	Base BaseModel `gorm:"embedded"`
	// Many-Many relations
	Devices []Device `gorm:"many2many:devicefamily_devices"`
}

// BaseModel returns the reference to the base model
func (entity *Devicefamily) BaseModel() *BaseModel {
	return &entity.Base
}

// BeforeCreate to validate and set default
func (entity *Devicefamily) BeforeCreate(tx *gorm.DB) error {
	return entity.Base.preCreate(tx, entity)
}

// Device -----------------------------------------------------------------
type Device struct {
	Base BaseModel `gorm:"embedded"`
	// Many-Many relations
	Devicefamilies []Devicefamily `gorm:"many2many:devicefamily_devices"`
}

// BaseModel returns the reference to the base model
func (entity *Device) BaseModel() *BaseModel {
	return &entity.Base
}

// BeforeCreate to validate and set default
func (entity *Device) BeforeCreate(tx *gorm.DB) error {
	return entity.Base.preCreate(tx, entity)
}
