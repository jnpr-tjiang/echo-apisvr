package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models/custom"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/utils"
	"github.com/mitchellh/mapstructure"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

//go:generate go run ../tools/codegen.go ../../schemas/model.yaml

type (
	// BaseModel - base database entity model
	BaseModel struct {
		ID          uuid.UUID               `gorm:"column:id;type:uuid;primary_key:<-create"`
		Name        string                  `gorm:"column:name;size:128;not null;<-:create"`
		DisplayName string                  `gorm:"column:display_name;size:128;not null"`
		ParentID    uuid.UUID               `gorm:"column:parent_id;type:uuid;<-:create"`
		ParentType  string                  `gorm:"column:parent_type;<-:create"`
		FQName      string                  `gorm:"column:fqname;not null;uniqueIndex;<:create"`
		Payload     datatypes.JSON          `gorm:"column:payload"`
		payload     *map[string]interface{} `gorm:"-"`
		refs        map[string]interface{}  `gorm:"-"`
	}

	// Entity is base interface for all models
	Entity interface {
		BaseModel() *BaseModel
		Find(db *gorm.DB, conds ...interface{}) ([]Entity, error)
	}

	// BaseRef - base ref entity model
	BaseRef struct {
		FromFQName string         `gorm:"column:from_fqname"`
		ToFQName   string         `gorm:"column:to_fqname"`
		Payload    datatypes.JSON `gorm:"column:payload"`
	}

	// RefEntity is the base interface for all refs
	RefEntity interface {
		BaseRef() *BaseRef
		Find(db *gorm.DB, conds ...interface{}) ([]RefEntity, error)
	}

	// ModelInfo contains entity model meta info
	ModelInfo struct {
		ParentTypes      []string
		ChildTypes       []string
		RefTypes         []string
		BackRefTypes     []string
		NormalizedFields []string
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
		obj := v()
		if entity, ok := obj.(Entity); ok {
			parentTypes, refTypes, fields, err := reflectEntity(entity)
			if err != nil {
				panic(err)
			}

			model := ModelInfo{}
			model.ParentTypes = parentTypes
			model.RefTypes = refTypes
			model.NormalizedFields = fields
			models[k] = &model
		}
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

func reflectEntity(entity Entity) (parentTypes []string, refTypes []string, normalizedFields []string, err error) {
	entityType := reflect.TypeOf(entity).Elem()

	baseField, ok := entityType.FieldByName("Base")
	if !ok {
		return []string{}, []string{}, []string{}, fmt.Errorf("Base field not found in entity: %s", entityType.Name())
	}
	if baseField.Type.Name() != "BaseModel" {
		return []string{}, []string{}, []string{}, fmt.Errorf("Base field must be BaseModel type: %s", entityType.Name())
	}
	if v, ok := baseField.Tag.Lookup("parentTypes"); ok {
		parentTypes = strings.Split(v, ",")
		for i := 0; i < len(parentTypes); i++ {
			if parentTypes[i] == "" {
				parentTypes = parentTypes[:len(parentTypes)-1]
			} else if utils.IndexOf(modelNames, parentTypes[i]) < 0 {
				return []string{}, []string{}, []string{}, fmt.Errorf("[%s] Invalid parent type: %s", entityType.Name(), parentTypes[i])
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
					} else {
						normalizedFields = append(normalizedFields, strings.ToLower(field.Name))
					}
				}
			}
		}
	}
	for i := 0; i < len(refTypes); i++ {
		if utils.IndexOf(modelNames, refTypes[i]) < 0 {
			return []string{}, []string{}, []string{}, fmt.Errorf("[%s] Invalid ref type: %s", entityType.Name(), refTypes[i])
		}
	}

	return parentTypes, refTypes, normalizedFields, nil
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

func findRefEntity(db *gorm.DB, dest interface{}, conds ...interface{}) ([]RefEntity, error) {
	if err := db.Find(dest, conds...).Error; err != nil {
		return []RefEntity{}, err
	}
	s := reflect.ValueOf(dest).Elem()
	if s.Kind() != reflect.Slice {
		return nil, fmt.Errorf("dest must be an slice")
	}

	entities := make([]RefEntity, s.Len(), s.Len())
	for i := 0; i < s.Len(); i++ {
		entities[i] = s.Index(i).Addr().Interface().(RefEntity)
	}
	return entities, nil
}

// NewEntity is the factory function to construct a new entity by type
func NewEntity(entityType string) (Entity, error) {
	c, ok := constructors[entityType]
	if !ok {
		return nil, fmt.Errorf("Invalid Entity type: " + entityType)
	}
	if entity, ok := c().(Entity); ok {
		return entity, nil
	}
	return nil, fmt.Errorf("Invalid Entity type: %s", entityType)
}

// NewRefEntity is the factory function to construct a new entity by type
func NewRefEntity(refEntityType string) (RefEntity, error) {
	c, ok := constructors[refEntityType]
	if !ok {
		return nil, fmt.Errorf("Invalid Entity type: " + refEntityType)
	}
	if refEntity, ok := c().(RefEntity); ok {
		return refEntity, nil
	}
	return nil, fmt.Errorf("Invalid RefEntity type: " + refEntityType)
}

// ModelNames returns names for all registered models
func ModelNames() []string {
	return modelNames
}

// GetModelInfo return entity model meta info
func GetModelInfo(entity Entity) (ModelInfo, bool) {
	m, ok := models[EntityType(entity)]
	return *m, ok
}

// GetIDByFQName queries DB for FQName by ID
func GetIDByFQName(tx *gorm.DB, fqname string, entityType string) (uuid.UUID, error) {
	var ids []uuid.UUID
	sql := fmt.Sprintf("select id from %s where fqname = ?", utils.Pluralize(entityType))
	tx.Raw(sql, fqname).Scan(&ids)
	if len(ids) == 0 {
		return EmptyUUID, fmt.Errorf("Parent id not found for fqname[%s]: %s", entityType, fqname)
	}
	return ids[0], nil
}

// GetFQNameByID queries DB for ID by FQName
func GetFQNameByID(tx *gorm.DB, id uuid.UUID, entityType string) (string, error) {
	sql := fmt.Sprintf("select fqname from %s where id = ?", utils.Pluralize(entityType))
	var fqname string
	tx.Raw(sql, id).Scan(&fqname)
	if fqname == "" {
		return fqname, fmt.Errorf("Failed to find parent obj: %s[%v]", entityType, id)
	}
	return fqname, nil
}

// PopulateEntity with json payload
func PopulateEntity(entity Entity, payload map[string]interface{}) (err error) {
	model, ok := GetModelInfo(entity)
	if !ok {
		return fmt.Errorf("Model info not found: %s", utils.TypeOf(entity))
	}

	// populate the BaseModel
	if uuidStr, ok := payload["uuid"]; ok {
		if entity.BaseModel().ID, err = uuid.Parse(uuidStr.(string)); err != nil {
			return err
		}
	}
	if name, ok := payload["name"]; ok {
		entity.BaseModel().Name = name.(string)
	}
	if displayName, ok := payload["display_name"]; ok {
		entity.BaseModel().DisplayName = displayName.(string)
	}
	if fqname, ok := payload["fq_name"]; ok {
		var s []string
		for _, v := range fqname.([]interface{}) {
			s = append(s, v.(string))
		}
		fqn := custom.FQName(s)
		if val, err := custom.FQName(fqn).Value(); err == nil {
			entity.BaseModel().FQName = val.(string)
		}
	}
	if parentType, ok := payload["parent_type"]; ok {
		entity.BaseModel().ParentType = parentType.(string)
	}
	if parentID, ok := payload["parent_uuid"]; ok {
		var pid uuid.UUID
		if pid, err = uuid.Parse(parentID.(string)); err != nil {
			return err
		}
		entity.BaseModel().ParentID = pid
	}

	// populate the normalized fields
	if len(model.NormalizedFields) > 0 {
		err = mapstructure.Decode(payload, &entity)
		if err != nil {
			return err
		}
	}

	// populate the refs
	entity.BaseModel().refs = make(map[string]interface{})
	for _, refType := range model.RefTypes {
		fieldName := refType + "_refs"
		if v, ok := payload[fieldName]; ok {
			entity.BaseModel().refs[refType] = v
		}
	}

	// remove refs and updatable normalized fields from payload
	payloadCleansing(&payload, model)
	entity.BaseModel().payload = &payload

	return err
}

func payloadCleansing(payload *map[string]interface{}, model ModelInfo) {
	// remove all duplicated fields that are captured in the BaseModel or
	// could be generated from the BaseModel
	delete(*payload, "uuid")
	delete(*payload, "uri")
	delete(*payload, "name")
	delete(*payload, "fq_name")
	delete(*payload, "display_name")
	delete(*payload, "parent_type")
	delete(*payload, "parent_uuid")
	delete(*payload, "parent_uri")

	// remove normalized fields
	for _, normalizedField := range model.NormalizedFields {
		delete(*payload, normalizedField)
	}

	// remove refs as back_refs from payload as they are stored in a separate table
	for _, refType := range model.RefTypes {
		fieldName := refType + "_refs"
		delete(*payload, fieldName)
	}
	for _, backRefType := range model.BackRefTypes {
		fieldName := backRefType + "_back_refs"
		delete(*payload, fieldName)
	}

	// remove child refs
	for _, childType := range model.ChildTypes {
		delete(*payload, childType)
	}
}

// EntityType return entity type
func EntityType(entity Entity) string {
	return strings.ToLower(utils.TypeOf(entity))
}

// SaveEntity to the database
func SaveEntity(db *gorm.DB, entity Entity) error {
	return db.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(entity).Error; err != nil {
			return err
		}
		for k, v := range entity.BaseModel().refs {
			if refs, ok := v.([]interface{}); ok {
				toType := k
				for _, refData := range refs {
					refEntity, err := NewRefEntity(EntityType(entity) + toType)
					if err != nil {
						return err
					}
					if err = loadRefEntity(tx, refEntity, entity, toType, refData); err != nil {
						return err
					}
					if err = tx.Create(refEntity).Error; err != nil {
						return err
					}
				}
			}
		}
		return err
	})
}

func zeroizeReadonlyFields(entity Entity) {
	entity.BaseModel().Name = ""
	entity.BaseModel().ParentID = uuid.UUID{}
	entity.BaseModel().ParentType = ""
	entity.BaseModel().FQName = ""
}

// UpdateEntity to the database
func UpdateEntity(db *gorm.DB, entity Entity) error {
	return db.Transaction(func(tx *gorm.DB) (err error) {
		// zeroize the readonly fields (such as fqname, etc) so that they are not updated by gorm
		zeroizeReadonlyFields(entity)

		if len(*entity.BaseModel().payload) == 0 {
			if err = tx.Updates(entity).Error; err != nil {
				return err
			}
		} else {
			currentEntity, _ := NewEntity(EntityType(entity))
			if err = tx.First(currentEntity).Error; err != nil {
				return err
			}
			var payloadJSON map[string]interface{}
			if err = json.Unmarshal(currentEntity.BaseModel().Payload, &payloadJSON); err != nil {
				return err
			}
			for k, v := range *entity.BaseModel().payload {
				payloadJSON[k] = v
			}
			entity.BaseModel().Payload, err = json.Marshal(payloadJSON)
			if err = tx.Updates(entity).Error; err != nil {
				return err
			}
		}

		// for k, v := range entity.BaseModel().refs {
		// 	if refs, ok := v.([]interface{}); ok {
		// 		toType := k
		// 		for _, refData := range refs {
		// 			refEntity, err := NewRefEntity(entityType + toType)
		// 			if err != nil {
		// 				return err
		// 			}
		// 			if err = loadRefEntity(tx, refEntity, entity, toType, refData); err != nil {
		// 				return err
		// 			}
		// 			if err = tx.Create(refEntity).Error; err != nil {
		// 				return err
		// 			}
		// 		}
		// 	}
		// }
		return err
	})
}

func loadRefEntity(tx *gorm.DB, refEntity RefEntity, fromEntity Entity, toType string, refData interface{}) (err error) {
	type RefStruct struct {
		UUID string
		To   []string
		Attr interface{}
	}
	var (
		refStruct            RefStruct = RefStruct{}
		fromID, toID         uuid.UUID
		fromFQName, toFQName string
		ok                   bool
	)
	if err = mapstructure.Decode(refData, &refStruct); err != nil {
		return err
	}

	fromID = fromEntity.BaseModel().ID
	fromFQName = fromEntity.BaseModel().FQName
	if refStruct.UUID != "" {
		toID, err = uuid.Parse(refStruct.UUID)
		if err != nil {
			return err
		}
		toFQName, err = GetFQNameByID(tx, toID, strings.ToLower(toType))
		if err != nil {
			return err
		}

	} else if len(refStruct.To) > 0 {
		val, err := custom.FQName(refStruct.To).Value()
		if err != nil {
			return err
		}
		toFQName, ok = val.(string)
		if !ok {
			return fmt.Errorf("Something is wrong")
		}
		toID, err = GetIDByFQName(tx, toFQName, strings.ToLower(toType))
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Invalid ref data: %v", refData)
	}

	refEntityMap := make(map[string]interface{})
	refEntityMap[utils.TypeOf(fromEntity)+"ID"] = fromID
	refEntityMap[strings.Title(toType)+"ID"] = toID
	base := make(map[string]interface{})
	base["FromFQName"] = fromFQName
	base["ToFQName"] = toFQName
	base["Payload"], err = json.Marshal(refStruct.Attr)
	if err != nil {
		return err
	}
	refEntityMap["Base"] = base
	if err = mapstructure.Decode(refEntityMap, refEntity); err != nil {
		return err
	}
	return nil
}

func (b *BaseModel) preCreate(tx *gorm.DB, obj Entity) (err error) {
	// name is mandatory field
	if b.Name == "" {
		return fmt.Errorf("Empty name not allowed")
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
	objType := EntityType(obj)
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
			parentFQN, err := custom.ParseParentFQName(b.FQName)
			if err != nil {
				return fmt.Errorf("Invalid fqname: %s", b.FQName)
			}
			if b.ParentID, err = GetIDByFQName(tx, parentFQN, b.ParentType); err != nil {
				return err
			}
		} else if b.FQName == "" && b.ParentID != EmptyUUID {
			parentFQName, err := GetFQNameByID(tx, b.ParentID, b.ParentType)
			if err != nil {
				return err
			}
			b.FQName = custom.ConstructFQName(parentFQName, b.Name)
		} else {
			return fmt.Errorf("Both fqname and parentID are not set")
		}
	}
	b.Payload, err = json.Marshal(*b.payload)
	return err
}
