package models

import (
	"fmt"

	"github.com/gertd/go-pluralize"
	"github.com/google/uuid"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/utils"
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

var (
	hierarchyMap map[string]string = make(map[string]string)
	pluralizer   *pluralize.Client = pluralize.NewClient()
)

func addHierarchy(parent interface{}, child interface{}) {
	hierarchyMap[utils.TypeOf(child)] = utils.TypeOf(parent)
}

// GetParentType returns object's parent type name
func GetParentType(modelObj interface{}) string {
	return hierarchyMap[utils.TypeOf(modelObj)]
}

// GetDbTableName returns model's corresponding db table name
func GetDbTableName(modelObj interface{}) string {
	return pluralizer.Plural(utils.TypeOf(modelObj))
}

func (b *BaseModel) preCreate(tx *gorm.DB, obj interface{}) (err error) {
	if b.Name == "" {
		return fmt.Errorf("Empty name not allow")
	}
	if b.ID == (uuid.UUID{}) {
		b.ID = uuid.New()
	}
	parentType := GetParentType(obj)
	if parentType == "" {
		b.ParentID = uuid.UUID{}
		b.FQName = fmt.Sprintf(`["%s"]`, b.Name)
	} else {
		if b.ParentID == (uuid.UUID{}) {
			return fmt.Errorf("Empty parent uuid not allow for %s", parentType)
		}
		sql := fmt.Sprintf("select fqname from %s where id = ?", pluralizer.Plural(parentType))
		var parentFQName string
		tx.Raw(sql, b.ParentID).Scan(&parentFQName)
		if parentFQName == "" {
			return fmt.Errorf("Failed to find parent obj: %s[%v]", parentType, b.ParentID)
		}
		b.FQName = fmt.Sprintf(`%s, "%s"]`, parentFQName[:len(parentFQName)-1], b.Name)
	}
	return nil
}
