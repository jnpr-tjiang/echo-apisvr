package models

import (
	"os"
	"testing"

	"github.com/jnpr-tjiang/echo-apisvr/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBaseModel_New(t *testing.T) {
	Init()
	entity, err := NewEntity("domain")
	if err != nil {
		t.Error(err)
		return
	}
	entityType := utils.TypeOf(entity)
	if entityType != "Domain" {
		t.Errorf("Expected value: Domain but got %s", entityType)
	}
}

func TestBaseModel_ModelNames(t *testing.T) {
	Init()
	names := ModelNames()
	want := []string{"domain", "project", "device", "devicefamily"}
	if len(names) != len(want) {
		t.Errorf("Expect %v but got %v", want, names)
	}
	for _, v := range names {
		if _, ok := utils.Find(want, v); !ok {
			t.Errorf("Expect %v but got %v", want, names)
		}
	}
}

func TestBaseModel_preCreate(t *testing.T) {
	os.Remove("./test.db")
	db, err := gorm.Open(sqlite.Open("./test.db"), &gorm.Config{})
	if err != nil {
		t.Errorf("Failed to open the database: %v", err)
		return
	}

	// initialize the database and create tables
	Init()
	MigrateDataModel(db)

	domain := Domain{
		Base: BaseModel{
			Name:    "default",
			Payload: datatypes.JSON([]byte(`{"display_name": "default", "system": {"serial": "SN1234", "mac":"ab:34:12:f3"}}`)),
		},
	}
	if err = db.Create(&domain).Error; err != nil {
		t.Errorf("Failed to create domain: %v", err)
		return
	}
	if domain.BaseModel().ID == EMPTY_UUID || domain.BaseModel().FQName != "[\"default\"]" {
		t.Errorf("Wrong ParentID")
		return
	}

	project := Project{
		Base: BaseModel{
			Name:     "juniper",
			ParentID: domain.BaseModel().ID,
		},
	}
	if err = db.Create(&project).Error; err != nil {
		t.Errorf("Failed to create project: %v", err)
		return
	}
	if project.BaseModel().ID == EMPTY_UUID ||
		project.BaseModel().FQName != "[\"default\", \"juniper\"]" ||
		project.BaseModel().ParentType != "domain" ||
		project.BaseModel().ParentID != domain.BaseModel().ID {
		t.Errorf("project ID is not set")
	}
}
