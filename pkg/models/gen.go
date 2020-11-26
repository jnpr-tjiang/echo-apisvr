package models

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var constructors map[string]func() interface{} = map[string]func() interface{}{
	"domain":              func() interface{} { return &Domain{} },
	"project":             func() interface{} { return &Project{} },
	"device":              func() interface{} { return &Device{} },
	"device_family":       func() interface{} { return &Device_family{} },
	"devicedevice_family": func() interface{} { return &DeviceDevice_family{} },
}

// Domain -----------------------------------------------------------------
type Domain struct {
	Base BaseModel `gorm:"embedded"`
}

// BaseModel returns the reference to the base model
func (entity *Domain) BaseModel() *BaseModel {
	return &entity.Base
}

// Find entities that meets certain conditions
func (entity *Domain) Find(db *gorm.DB, conds ...interface{}) ([]Entity, error) {
	var entities []Domain
	return findEntity(db, &entities, conds...)
}

// BeforeCreate to validate and set default
func (entity *Domain) BeforeCreate(tx *gorm.DB) error {
	return entity.Base.preCreate(tx, entity)
}

// Project -----------------------------------------------------------------
type Project struct {
	Base BaseModel `gorm:"embedded" parentTypes:"domain"`
}

// BaseModel returns the reference to the base model
func (entity *Project) BaseModel() *BaseModel {
	return &entity.Base
}

// Find entities that meets certain conditions
func (entity *Project) Find(db *gorm.DB, conds ...interface{}) ([]Entity, error) {
	var entities []Project
	return findEntity(db, &entities, conds...)
}

// BeforeCreate to validate and set default
func (entity *Project) BeforeCreate(tx *gorm.DB) error {
	return entity.Base.preCreate(tx, entity)
}

// Device_family -----------------------------------------------------------------
type Device_family struct {
	Base BaseModel `gorm:"embedded" parentTypes:"project"`
}

// BaseModel returns the reference to the base model
func (entity *Device_family) BaseModel() *BaseModel {
	return &entity.Base
}

// Find entities that meets certain conditions
func (entity *Device_family) Find(db *gorm.DB, conds ...interface{}) ([]Entity, error) {
	var entities []Device_family
	return findEntity(db, &entities, conds...)
}

// BeforeCreate to validate and set default
func (entity *Device_family) BeforeCreate(tx *gorm.DB) error {
	return entity.Base.preCreate(tx, entity)
}

// Device -----------------------------------------------------------------
type Device struct {
	Base BaseModel `gorm:"embedded" parentTypes:"domain,project"`
	// fields
	Connection_type string `gorm:"column:connction_type"`
	// Many-Many relations
	Device_families []Device_family `gorm:"many2many:device_device_families"`
}

// BaseModel returns the reference to the base model
func (entity *Device) BaseModel() *BaseModel {
	return &entity.Base
}

// Find entities that meets certain conditions
func (entity *Device) Find(db *gorm.DB, conds ...interface{}) ([]Entity, error) {
	var entities []Device
	return findEntity(db, &entities, conds...)
}

// BeforeCreate to validate and set default
func (entity *Device) BeforeCreate(tx *gorm.DB) error {
	return entity.Base.preCreate(tx, entity)
}

// DeviceDevice_family -----------------------------------------------------------
type DeviceDevice_family struct {
	Base            BaseRef `gorm:"embedded"`
	DeviceID        uuid.UUID
	Device_familyID uuid.UUID
}

// BaseRef returns the reference to the BaseRef model
func (entity *DeviceDevice_family) BaseRef() *BaseRef {
	return nil
}

// Find entities that meets certain conditions
func (entity *DeviceDevice_family) Find(db *gorm.DB, conds ...interface{}) ([]RefEntity, error) {
	var entities []DeviceDevice_family
	return findRefEntity(db, &entities, conds...)
}
