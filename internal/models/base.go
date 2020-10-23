package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// BaseModel - base database entity model
type BaseModel struct {
	ID       uuid.UUID      `gorm:"column:id;type:uuid;primary_key"`
	Name     string         `gorm:"column:name;size:128;not null;<-:create"`
	ParentID uuid.UUID      `gorm:"column:parent_id;type:uuid"`
	FQName   string         `gorm:"column:fqname;not null;uniqueIndex"`
	Payload  datatypes.JSON `gorm:"column:payload"`
}
