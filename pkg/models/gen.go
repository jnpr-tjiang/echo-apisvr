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
		allowedParentTypes: []string{"project", "domain"},
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
	Devices  []Device  `gorm:"foreignKey:ParentID;references:ID"`
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
	Base BaseModel `gorm:"embedded"`
	// Has-Many relations
	Devices        []Device       `gorm:"foreignKey:ParentID;references:ID"`
	DeviceFamilies []Devicefamily `gorm:"foreignKey:ParentID;references:ID"`
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
	Base BaseModel `gorm:"embedded"`
	// Many-Many relations
	Devices []Device `gorm:"many2many:devicefamily_devices"`
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
	Base BaseModel `gorm:"embedded"`
	// Many-Many relations
	Devicefamilies []Devicefamily `gorm:"many2many:devicefamily_devices"`
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
