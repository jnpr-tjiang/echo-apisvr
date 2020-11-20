package models

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var constructors map[string]func() interface{} = map[string]func() interface{}{
	"domain":             func() interface{} { return &Domain{} },
	"project":            func() interface{} { return &Project{} },
	"device":             func() interface{} { return &Device{} },
	"devicefamily":       func() interface{} { return &Devicefamily{} },
	"devicedevicefamily": func() interface{} { return &DeviceDevicefamily{} },
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

// Devicefamily -----------------------------------------------------------------
type Devicefamily struct {
	Base BaseModel `gorm:"embedded" parentTypes:"project"`
}

// BaseModel returns the reference to the base model
func (entity *Devicefamily) BaseModel() *BaseModel {
	return &entity.Base
}

// Find entities that meets certain conditions
func (entity *Devicefamily) Find(db *gorm.DB, conds ...interface{}) ([]Entity, error) {
	var entities []Devicefamily
	return findEntity(db, &entities, conds...)
}

// BeforeCreate to validate and set default
func (entity *Devicefamily) BeforeCreate(tx *gorm.DB) error {
	return entity.Base.preCreate(tx, entity)
}

// Device -----------------------------------------------------------------
type Device struct {
	Base BaseModel `gorm:"embedded" parentTypes:"domain,project"`
	// Many-Many relations
	Devicefamilies []Devicefamily `gorm:"many2many:device_devicefamilies"`
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

// DeviceDevicefamily -----------------------------------------------------------
type DeviceDevicefamily struct {
	Base           BaseRef `gorm:"embedded"`
	DeviceID       uuid.UUID
	DevicefamilyID uuid.UUID
}

// BaseRef returns the reference to the BaseRef model
func (entity *DeviceDevicefamily) BaseRef() *BaseRef {
	return nil
}

// Find entities that meets certain conditions
func (entity *DeviceDevicefamily) Find(db *gorm.DB, conds ...interface{}) ([]RefEntity, error) {
	var entities []DeviceDevicefamily
	return findRefEntity(db, &entities, conds...)
}
