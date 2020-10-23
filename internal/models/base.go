package models

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BaseModel - base database entity model
type BaseModel struct {
	ID       uuid.UUID      `gorm:"column:id;type:uuid;primary_key"`
	Name     string         `gorm:"column:name;size:128;not null;<-:create"`
	ParentID uuid.UUID      `gorm:"column:parent_id;type:uuid"`
	FQName   string         `gorm:"column:fqname;not null;uniqueIndex"`
	Payload  datatypes.JSON `gorm:"column:payload"`
}

func (b *BaseModel) preCreate(tx *gorm.DB, parentType string) (err error) {
	if b.Name == "" {
		return fmt.Errorf("Empty name not allow")
	}
	if b.ID == (uuid.UUID{}) {
		b.ID = uuid.New()
	}
	if parentType == "" {
		b.ParentID = uuid.UUID{}
		b.FQName = fmt.Sprintf(`["%s"]`, b.Name)
	} else {
		if b.ParentID == (uuid.UUID{}) {
			return fmt.Errorf("Empty parent uuid not allow for %s", parentType)
		}
		sql := fmt.Sprintf("select fqname from %ss where id = ?", parentType)
		var parentFQName string
		if err = tx.Raw(sql, b.ParentID).Scan(&parentFQName).Error; err != nil {
			return fmt.Errorf("Invalid parent uuid: %v", b.ParentID)
		}
		b.FQName = fmt.Sprintf(`%s, "%s"]`, parentFQName[:len(parentFQName)-1], b.Name)
	}
	return nil
}
