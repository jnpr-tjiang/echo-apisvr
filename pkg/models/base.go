package models

import (
	"encoding/json"
	"fmt"
	"reflect"
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
		ID          uuid.UUID               `gorm:"column:id;type:uuid;primary_key"`
		Name        string                  `gorm:"column:name;size:128;not null;<-:create"`
		DisplayName string                  `gorm:"column:display_name;size:128;not null"`
		ParentID    uuid.UUID               `gorm:"column:parent_id;type:uuid"`
		ParentType  string                  `gorm:"column:parent_type"`
		FQName      string                  `gorm:"column:fqname;not null;uniqueIndex"`
		Payload     datatypes.JSON          `gorm:"column:payload"`
		JSON        *map[string]interface{} `gorm:"-"`
	}

	// Entity is base interface for all models
	Entity interface {
		BaseModel() *BaseModel
		Find(db *gorm.DB, conds ...interface{}) ([]Entity, error)
	}

	// ModelInfo contains entity model meta info
	ModelInfo struct {
		ParentTypes  []string
		ChildTypes   []string
		RefTypes     []string
		BackRefTypes []string
	}
)

var (
	// EmptyUUID for empty UUID
	EmptyUUID  uuid.UUID             = uuid.UUID{}
	models     map[string]*ModelInfo = make(map[string]*ModelInfo)
	modelNames []string
)

func init() {
	modelNames = make([]string, len(constructors))
	i := 0
	for k := range constructors {
		modelNames[i] = k
		i++
	}

	for k, v := range constructors {
		entity := v()
		parentTypes, refTypes, err := reflectEntity(entity)
		if err != nil {
			panic(err)
		}

		model := ModelInfo{}
		model.ParentTypes = parentTypes
		model.RefTypes = refTypes
		models[k] = &model
	}

	for entityType, model := range models {
		for _, parentType := range model.ParentTypes {
			parentModel, ok := models[parentType]
			if !ok {
				panic(fmt.Sprintf("[%s] Invalid parent type: %s", entityType, parentType))
			}
			parentModel.ChildTypes = append(parentModel.ChildTypes, entityType)
		}
		for _, refType := range model.RefTypes {
			refModel, ok := models[refType]
			if !ok {
				panic(fmt.Sprintf("[%s] Invalid ref type: %s", entityType, refType))
			}
			refModel.BackRefTypes = append(refModel.BackRefTypes, entityType)
		}
	}
}

func reflectEntity(entity Entity) (parentTypes []string, refTypes []string, err error) {
	entityType := reflect.TypeOf(entity).Elem()

	baseField, ok := entityType.FieldByName("Base")
	if !ok {
		return []string{}, []string{}, fmt.Errorf("Base field not found in entity: %s", entityType.Name())
	}
	if baseField.Type.Name() != "BaseModel" {
		return []string{}, []string{}, fmt.Errorf("Base field must be BaseModel type: %s", entityType.Name())
	}
	if v, ok := baseField.Tag.Lookup("parentTypes"); ok {
		parentTypes = strings.Split(v, ",")
		for i := 0; i < len(parentTypes); i++ {
			if utils.IndexOf(modelNames, parentTypes[i]) < 0 {
				return []string{}, []string{}, fmt.Errorf("[%s] Invalid parent type: %s", entityType.Name(), parentTypes[i])
			}
		}
	}

	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		if field.Type.Name() != "BaseModel" {
			if v, ok := field.Tag.Lookup("gorm"); ok {
				gormDirectives := strings.Split(v, ";")
				for _, directive := range gormDirectives {
					if strings.Index(strings.Trim(directive, " "), "many2many") >= 0 {
						refTypes = append(refTypes, utils.Singularize(strings.ToLower(field.Name)))
					}
				}
			}
		}
	}
	for i := 0; i < len(refTypes); i++ {
		if utils.IndexOf(modelNames, refTypes[i]) < 0 {
			return []string{}, []string{}, fmt.Errorf("[%s] Invalid ref type: %s", entityType.Name(), refTypes[i])
		}
	}

	return parentTypes, refTypes, nil
}

func findEntity(db *gorm.DB, dest interface{}, conds ...interface{}) ([]Entity, error) {
	if err := db.Find(dest, conds...).Error; err != nil {
		return []Entity{}, err
	}
	s := reflect.ValueOf(dest).Elem()
	if s.Kind() != reflect.Slice {
		return nil, fmt.Errorf("dest must be an slice")
	}

	entities := make([]Entity, s.Len(), s.Len())
	for i := 0; i < s.Len(); i++ {
		entities[i] = s.Index(i).Addr().Interface().(Entity)
	}
	return entities, nil
}

// NewEntity is the factory function to construct a new entity by type
func NewEntity(entityType string) (Entity, error) {
	c, ok := constructors[entityType]
	if !ok {
		return nil, fmt.Errorf("Invalid Entity type: " + entityType)
	}
	return c(), nil
}

// ModelNames returns names for all registered models
func ModelNames() []string {
	return modelNames
}

// GetModelInfo return entity model meta info
func GetModelInfo(entity Entity) (ModelInfo, bool) {
	m, ok := models[strings.ToLower(utils.TypeOf(entity))]
	return *m, ok
}

func (b *BaseModel) preCreate(tx *gorm.DB, obj Entity) (err error) {
	// name is mandatory field
	if b.Name == "" {
		return fmt.Errorf("Empty name not allow")
	}

	// auto set the display name if not set
	if b.DisplayName == "" {
		b.DisplayName = b.Name
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
	if b.ParentType == "" && len(m.ParentTypes) > 0 {
		b.ParentType = m.ParentTypes[0]
	}
	if idx := utils.IndexOf(m.ParentTypes, b.ParentType); b.ParentType != "" && idx < 0 {
		return fmt.Errorf("Invalid parent type: %s", b.ParentType)
	}

	// auto fill fqname or ParentID
	if b.ParentType == "" {
		b.ParentID = EmptyUUID
		b.FQName = fmt.Sprintf(`["%s"]`, b.Name)
	} else {
		// if both FQName and parentID are not empty, FQName takes the prededence
		if b.FQName != "" && b.ParentID != EmptyUUID {
			b.ParentID = EmptyUUID
		}

		if b.FQName != "" && b.ParentID == EmptyUUID {
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
		} else if b.FQName == "" && b.ParentID != EmptyUUID {
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
	b.constructPayload(obj)
	return nil
}

func (b *BaseModel) constructPayload(obj Entity) (err error) {
	idstr := b.ID.String()
	objType := strings.ToLower(utils.TypeOf(obj))
	(*b.JSON)["uuid"] = idstr
	if b.ParentType != "" {
		(*b.JSON)["parent_type"] = b.ParentType
		(*b.JSON)["parent_uuid"] = b.ParentID.String()
		(*b.JSON)["parent_uri"] = fmt.Sprintf("/%s/%s", b.ParentType, b.ParentID.String())
	}
	(*b.JSON)["uri"] = fmt.Sprintf("/%s/%s", objType, idstr)
	(*b.JSON)["display_name"] = b.DisplayName

	var fqname []string
	if err = json.Unmarshal([]byte(b.FQName), &fqname); err != nil {
		return err
	}
	(*b.JSON)["fq_name"] = fqname

	b.Payload, err = json.Marshal(*b.JSON)
	return err
}
