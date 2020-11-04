package models

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models/custom"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type (
	// BaseModel - base database entity model
	BaseModel struct {
		ID          uuid.UUID      `gorm:"column:id;type:uuid;primary_key"`
		Name        string         `gorm:"column:name;size:128;not null;<-:create"`
		DisplayName string         `gorm:"column:name;size:128;not null;<-:create"`
		ParentID    uuid.UUID      `gorm:"column:parent_id;type:uuid"`
		ParentType  string         `gorm:"column:parent_type`
		FQName      string         `gorm:"column:fqname;not null;uniqueIndex"`
		Payload     datatypes.JSON `gorm:"column:payload"`
	}

	// Entity is base interface for all models
	Entity interface {
		BaseModel() *BaseModel
	}

	modelInfo struct {
		allowedParentTypes []string
		constructor        func() Entity
	}
)

var (
	EMPTY_UUID uuid.UUID            = uuid.UUID{}
	models     map[string]modelInfo = make(map[string]modelInfo)
)

func register(modelType string, info modelInfo) {
	models[modelType] = info
}

// NewEntity is the factory function to construct a new entity by type
func NewEntity(entityType string) (Entity, error) {
	m, ok := models[entityType]
	if !ok {
		return nil, fmt.Errorf("Invalid Entity type: " + entityType)
	}
	if m.constructor == nil {
		return nil, fmt.Errorf("Entity constructor not found: " + entityType)
	}
	return m.constructor(), nil
}

// ModelNames returns names for all registered models
func ModelNames() []string {
	names := make([]string, len(models))
	i := 0
	for k := range models {
		names[i] = k
		i++
	}
	return names
}

func (b *BaseModel) preCreate(tx *gorm.DB, obj Entity) (err error) {
	// name is mandatory field
	if b.Name == "" {
		return fmt.Errorf("Empty name not allow")
	}

	// auto set the ID if not set
	if b.ID == (uuid.UUID{}) {
		b.ID = uuid.New()
	}

	// set or validate the parent type
	objType := strings.ToLower(utils.TypeOf(obj))
	m, ok := models[objType]
	if !ok {
		return fmt.Errorf("Model not supported: %s", objType)
	}
	if b.ParentType == "" && len(m.allowedParentTypes) > 0 {
		b.ParentType = m.allowedParentTypes[0]
	}
	if _, ok = utils.Find(m.allowedParentTypes, b.ParentType); b.ParentType != "" && !ok {
		return fmt.Errorf("Invalid parent type: %s", b.ParentType)
	}

	// auto fill fqname or ParentID
	if b.ParentType == "" {
		b.ParentID = EMPTY_UUID
		b.FQName = fmt.Sprintf(`["%s"]`, b.Name)
	} else {
		// if both FQName and parentID are not empty, FQName takes the prededence
		if b.FQName != "" && b.ParentID != EMPTY_UUID {
			b.ParentID = EMPTY_UUID
		}

		if b.FQName != "" && b.ParentID == EMPTY_UUID {
			sql := fmt.Sprintf("select id from %s where fqname = ?", utils.Pluralize(b.ParentType))
			parentFQN, err := custom.ParseParentFQName(b.FQName)
			if err != nil {
				return fmt.Errorf("Invalid fqname: %s", b.FQName)
			}
			var ids []uuid.UUID
			tx.Raw(sql, parentFQN).Scan(&ids)
			if len(ids) == 0 {
				return fmt.Errorf("Parent id not found for fqname[%s]: %s", b.ParentType, parentFQN)
			}
			b.ParentID = ids[0]
		} else if b.FQName == "" && b.ParentID != EMPTY_UUID {
			sql := fmt.Sprintf("select fqname from %s where id = ?", utils.Pluralize(b.ParentType))
			var parentFQName string
			tx.Raw(sql, b.ParentID).Scan(&parentFQName)
			if parentFQName == "" {
				return fmt.Errorf("Failed to find parent obj: %s[%v]", b.ParentType, b.ParentID)
			}
			b.FQName = custom.ConstructFQName(parentFQName, b.Name)
		} else {
			return fmt.Errorf("Both fqname and parentID are not set")
		}
	}
	return nil
}
