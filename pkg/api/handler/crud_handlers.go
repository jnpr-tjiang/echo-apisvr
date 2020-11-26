package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	guuid "github.com/gofrs/uuid"
	"github.com/google/uuid"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/database"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models/custom"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/utils"
	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PayloadCfg instructs how to build the payload
type PayloadCfg struct {
	ShowDetails  bool
	StrictFields bool
	Fields       []string
	ShowRefs     bool
	ShowBackRefs bool
	ShowChildren bool
}

func fieldsToShow(cfgFields []string, show bool, types []string) (fields []string) {
	for _, refType := range types {
		if utils.IndexOf(cfgFields, refType) >= 0 {
			fields = append(fields, refType)
		}
	}
	if show && len(fields) == 0 {
		fields = types
	}
	return fields
}

// RefFieldsToShow returns all the ref fields to include in the payload
func (cfg *PayloadCfg) RefFieldsToShow(modelInfo models.ModelInfo) (fields []string) {
	return fieldsToShow(cfg.Fields, cfg.ShowRefs, modelInfo.RefTypes)
}

// BackRefFieldsToShow returns all the backref fields to include in the payload
func (cfg *PayloadCfg) BackRefFieldsToShow(modelInfo models.ModelInfo) (fields []string) {
	return fieldsToShow(cfg.Fields, cfg.ShowBackRefs, modelInfo.BackRefTypes)
}

// ChildFieldsToShow returns all the child fields to include in the payload
func (cfg *PayloadCfg) ChildFieldsToShow(modelInfo models.ModelInfo) (fields []string) {
	return fieldsToShow(cfg.Fields, cfg.ShowChildren, modelInfo.ChildTypes)
}

func getPayloadCfg(c echo.Context) PayloadCfg {
	cfg := PayloadCfg{
		ShowDetails:  true,
		StrictFields: false,
		ShowRefs:     false,
		ShowBackRefs: false,
		ShowChildren: false,
	}

	var queryParams map[string][]string
	queryParams = c.QueryParams()
	if strings.Index(c.Path(), ":id") < 0 {
		if details, ok := queryParams["detail"]; ok && len(details) == 1 && details[0] == "true" {
			cfg.ShowDetails = true
		} else {
			cfg.ShowDetails = false
		}
	} else {
		if _, ok := queryParams["strict_fields"]; ok {
			cfg.StrictFields = true
		}
		if fields, ok := queryParams["fields"]; ok {
			cfg.Fields = fields
			// TODO: add logic to handle ref and child fields
		} else {
			if excludeRefs, ok := queryParams["exclude_refs"]; ok && len(excludeRefs) == 1 && excludeRefs[0] == "false" {
				cfg.ShowRefs = true
			}
			if excludeBackRefs, ok := queryParams["exclude_back_refs"]; ok && len(excludeBackRefs) == 1 && excludeBackRefs[0] == "false" {
				cfg.ShowBackRefs = true
			}
			if excludeChildren, ok := queryParams["exclude_children"]; ok && len(excludeChildren) == 1 && excludeChildren[0] == "false" {
				cfg.ShowChildren = true
			}
		}
	}
	return cfg
}

func filterFields(payload []byte, fields []string) (map[string]interface{}, error) {
	var jsonPayload map[string]interface{}
	if len(fields) == 0 {
		return jsonPayload, fmt.Errorf("Empty field list not allowed")
	}
	fields = append(
		fields, "name", "display_name", "fq_name", "uuid", "uri",
		"parent_uuid", "parent_uri", "parent_type")

	if err := json.Unmarshal(payload, &jsonPayload); err != nil {
		return jsonPayload, err
	}
	filteredPayload := make(map[string]interface{})
	for _, field := range fields {
		if v, ok := jsonPayload[field]; ok {
			filteredPayload[field] = v
		}
	}
	return filteredPayload, nil
}

func addBaseModelToPayload(entity models.Entity) {
	uuidStr := entity.BaseModel().ID.String()
	entityType := strings.ToLower(utils.TypeOf(entity))
	var str string
	if entity.BaseModel().ParentType == "" {
		str = fmt.Sprintf(
			`,"uuid":"%s","uri":"/%s/%s","name":"%s","fq_name":%s,"display_name":"%s"}`,
			uuidStr, entityType, uuidStr, entity.BaseModel().Name,
			entity.BaseModel().FQName, entity.BaseModel().DisplayName)
	} else {
		str = fmt.Sprintf(
			`,"uuid":"%s","uri":"/%s/%s","name":"%s","fq_name":%s,"display_name":"%s","parent_type":"%s","parent_uuid":"%s","parent_uri":"/%s/%s"}`,
			uuidStr, entityType, uuidStr, entity.BaseModel().Name, entity.BaseModel().FQName,
			entity.BaseModel().DisplayName, entity.BaseModel().ParentType, entity.BaseModel().ParentID,
			entity.BaseModel().ParentType, entity.BaseModel().ParentID)
	}
	payload := entity.BaseModel().Payload
	if len(payload) == 2 { // empty json. This looks like a hack. TODO: refactor this method to models?
		payload = []byte(str)
		payload[0] = '{'
	} else {
		payload = append(payload[:len(payload)-1], str...)
	}
	entity.BaseModel().Payload = payload
}

func addNormalizedFieldsToPayload(entity models.Entity, normalizedFieldNames []string) (err error) {
	var (
		fieldsMap map[string]interface{}
		entityMap map[string]interface{}
		payload   datatypes.JSON = entity.BaseModel().Payload
	)

	if len(normalizedFieldNames) == 0 {
		goto FINALLY
	}

	entity.BaseModel().Payload = datatypes.JSON{}
	entityMap = make(map[string]interface{})
	err = mapstructure.Decode(entity, &entityMap)
	if err != nil {
		goto FINALLY
	}
	fieldsMap = make(map[string]interface{})
	for _, fieldName := range normalizedFieldNames {
		if v, ok := entityMap[strings.Title(fieldName)]; ok {
			fieldsMap[fieldName] = v
		}
	}
	if len(fieldsMap) > 0 {
		if str, err := json.Marshal(fieldsMap); err == nil {
			payload[len(payload)-1] = ','
			payload = append(payload, str[1:]...)
		}
	}
	payload = append(payload[:len(payload)-1],
		[]byte(fmt.Sprintf(`,"display_name":"%s"}`, entity.BaseModel().DisplayName))...)
FINALLY:
	entity.BaseModel().Payload = payload
	return err
}

func buildEntityPayload(db *gorm.DB, entity models.Entity, cfg PayloadCfg) (payload []byte, err error) {
	if !cfg.ShowDetails {
		uuid := entity.BaseModel().ID.String()
		payload := fmt.Sprintf(
			`{"fq_name":%s,"uuid":"%s","uri":"/%s/%s"}`,
			entity.BaseModel().FQName, uuid, strings.ToLower(utils.TypeOf(entity)), uuid)
		return []byte(payload), nil
	}

	modelInfo, ok := models.GetModelInfo(entity)
	if !ok {
		return []byte{}, fmt.Errorf("Failed to get model info for entity: %v", entity)
	}
	addBaseModelToPayload(entity)
	if err = addNormalizedFieldsToPayload(entity, modelInfo.NormalizedFields); err != nil {
		return []byte{}, err
	}

	var filteredPayload map[string]interface{}
	if cfg.StrictFields && len(cfg.Fields) > 0 {
		if filteredPayload, err = filterFields(entity.BaseModel().Payload, cfg.Fields); err != nil {
			return []byte{}, err
		}
	}

	fields := make(map[string]interface{})
	refFieldsToShow := cfg.RefFieldsToShow(modelInfo)
	if len(refFieldsToShow) > 0 {
		refFields, err := buildRefFields(db, entity, refFieldsToShow, false)
		if err != nil {
			return []byte{}, err
		}
		for k, v := range refFields {
			fields[k] = v
		}
	}
	backRefFieldsToShow := cfg.BackRefFieldsToShow(modelInfo)
	if len(backRefFieldsToShow) > 0 {
		backRefFields, err := buildRefFields(db, entity, backRefFieldsToShow, true)
		if err != nil {
			return []byte{}, err
		}
		for k, v := range backRefFields {
			fields[k] = v
		}
	}
	childFieldsToShow := cfg.ChildFieldsToShow(modelInfo)
	if len(childFieldsToShow) > 0 {
		childFields, err := buildChildrenFields(db, entity, childFieldsToShow)
		if err != nil {
			return []byte{}, err
		}
		for k, v := range childFields {
			fields[k] = v
		}
	}

	if filteredPayload != nil && len(filteredPayload) > 0 {
		for k, v := range fields {
			filteredPayload[k] = v
		}
		return json.Marshal(filteredPayload)
	} else {
		payload := entity.BaseModel().Payload
		if len(fields) > 0 {
			payload[len(payload)-1] = ','
			fieldJSON, err := json.Marshal(fields)
			if err != nil {
				return []byte{}, err
			}
			payload = append(payload, fieldJSON[1:]...)
		}
		return payload, nil
	}
}

func buildChildrenFields(db *gorm.DB, entity models.Entity, childFieldsToShow []string) (map[string]interface{}, error) {
	childFields := make(map[string]interface{})
	for _, childField := range childFieldsToShow {
		if childEntity, err := models.NewEntity(childField); err == nil {
			childEntities, err := childEntity.Find(
				db.Select("ID", "fqname"), "parent_id = ? AND parent_type = ?",
				entity.BaseModel().ID, strings.ToLower(utils.TypeOf(entity)))
			if err != nil {
				return childFields, err
			}
			if len(childEntities) > 0 {
				childFields[utils.Pluralize(childField)] = buildChildRefs(childEntities)
			}
		}
	}
	return childFields, nil
}

func buildRefFields(db *gorm.DB, entity models.Entity, refFieldsToShow []string, isBackRef bool) (map[string]interface{}, error) {
	var (
		refEntityType string
	)
	refFields := make(map[string]interface{})
	for _, refField := range refFieldsToShow {
		entityType := utils.TypeOf(entity)
		if !isBackRef {
			refEntityType = strings.ToLower(entityType) + refField
		} else {
			refEntityType = refField + strings.ToLower(entityType)
		}
		if refEntity, err := models.NewRefEntity(refEntityType); err == nil {
			refEntities, err := refEntity.Find(db, fmt.Sprintf("%s_id = ?", entityType), entity.BaseModel().ID)
			if err != nil {
				return refFields, err
			}
			if len(refEntities) > 0 {
				var refFieldName string
				if isBackRef {
					refFieldName = refField + "_back_refs"
				} else {
					refFieldName = refField + "_refs"
				}
				refFields[refFieldName], err = buildRefs(refField, refEntities, isBackRef)
				if err != nil {
					return refFields, err
				}
			}
		}
	}
	return refFields, nil
}

func buildChildRefs(entities []models.Entity) []interface{} {
	refs := make([]interface{}, len(entities))
	for i, entity := range entities {
		uuidStr := entity.BaseModel().ID.String()
		entityType := strings.ToLower(utils.TypeOf(entity))
		jsonStr := fmt.Sprintf(`{"to": %s, "uri": "/%s/%s", "uuid": "%s"}`, entity.BaseModel().FQName, entityType, uuidStr, uuidStr)
		var refJSON interface{}
		json.Unmarshal([]byte(jsonStr), &refJSON)
		refs[i] = refJSON
	}
	return refs
}

func buildRefs(refField string, refEntities []models.RefEntity, isBackRef bool) ([]interface{}, error) {
	var (
		ok        bool
		fqname    custom.FQName
		fqnameStr string
		refUUID   guuid.UUID
		v         interface{}
	)
	refs := make([]interface{}, len(refEntities))
	for i, refEntity := range refEntities {
		refEntityMap := make(map[string]interface{})
		err := mapstructure.Decode(refEntity, &refEntityMap)
		if err != nil {
			return refs, err
		}

		ref := make(map[string]interface{})
		refFieldType := strings.Title(refField)
		if ref["uuid"], ok = refEntityMap[refFieldType+"ID"]; !ok {
			return refs, fmt.Errorf(fmt.Sprintf("Filed %sID not exist in %v", refFieldType, refEntityMap))
		}
		v, _ = ref["uuid"]
		refUUID, ok = v.(guuid.UUID)
		if !ok {
			return refs, fmt.Errorf("Invalid uuid type: %v", v)
		}
		ref["uuid"] = refUUID.String()
		ref["uri"] = fmt.Sprintf("/%s/%s", refField, refUUID.String())

		v, ok := refEntityMap["Base"]
		if !ok {
			return refs, fmt.Errorf("Field 'Base' not exist in %v", refEntityMap)
		}
		base, ok := v.(map[string]interface{})
		if !ok {
			return refs, fmt.Errorf("Invalid 'Base' field type")
		}
		var fieldName string
		if !isBackRef {
			fieldName = "ToFQName"
		} else {
			fieldName = "FromFQName"
		}
		v, ok = base[fieldName]
		fqnameStr, ok = v.(string)
		if !ok {
			return refs, fmt.Errorf("Field '%s' is not a string type: %v", fieldName, refEntityMap)
		}
		err = fqname.Scan(fqnameStr)
		if err != nil {
			return refs, err
		}
		ref["to"] = fqname
		data, ok := base["Payload"]
		if ok {
			if payload, ok := data.(datatypes.JSON); ok {
				err = json.Unmarshal(payload, &v)
				if err != nil {
					return refs, err
				}
				ref["attr"] = v
			}
		}
		refs[i] = ref
	}
	return refs, nil
}

// ModelCreateHandler for request to create a model entity
func ModelCreateHandler(c echo.Context) error {
	// get validated payload from context
	validationErrMsg := c.Get("validationErrors")
	if validationErrMsg != "" {
		return c.String(http.StatusBadRequest, validationErrMsg.(string))
	}
	p := c.Get("validatedPayload")
	if p == nil {
		return fmt.Errorf("No validated payload found in the context")
	}
	payload := p.(map[string]interface{})

	// create the entity
	entityType := strings.Split(c.Path(), "/")[1]
	entity, err := models.NewEntity(entityType)
	if err != nil {
		return err
	}
	err = models.PopulateEntity(entity, payload)
	if err != nil {
		return err
	}

	// save the entity to database
	db := database.GormDB()
	if err = models.SaveEntity(db, entity); err != nil {
		return err
	}

	return c.String(http.StatusCreated, fmt.Sprintf("%s", entity.BaseModel().ID))
}

// ModelGetAllHandler for request to get all model entities
func ModelGetAllHandler(c echo.Context) error {
	entityType := strings.Split(c.Path(), "/")[1]
	entity, err := models.NewEntity(entityType)
	if err != nil {
		return err
	}

	db := database.GormDB()
	entities, err := entity.Find(db)
	if err != nil {
		return err
	}
	body := []byte(fmt.Sprintf(`{"total": %d, "%s": [`, len(entities), entityType))
	for i, v := range entities {
		payload, err := buildEntityPayload(db, v, getPayloadCfg(c))
		if err != nil {
			return err
		}
		body = append(body, payload...)
		if (i + 1) != len(entities) {
			body = append(body, ","...)
		}
	}
	body = append(body, []byte("]}")...)
	return c.Blob(http.StatusOK, echo.MIMEApplicationJSON, body)
}

// ModelGetHandler for request to get an model entity by id
func ModelGetHandler(c echo.Context) error {
	entityType := strings.Split(c.Path(), "/")[1]
	entity, err := models.NewEntity(entityType)
	if err != nil {
		return err
	}
	uuid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return err
	}
	db := database.GormDB()
	if err = db.First(entity, uuid).Error; err != nil {
		return err
	}

	payload, err := buildEntityPayload(db, entity, getPayloadCfg(c))
	if err != nil {
		return err
	}
	body := []byte(fmt.Sprintf(`{"%s":`, entityType))
	body = append(body, payload...)
	body = append(body, []byte("}")...)
	return c.Blob(http.StatusOK, echo.MIMEApplicationJSON, body)
}

// ModelUpdateHandler for request to update a model entity
func ModelUpdateHandler(c echo.Context) error {
	// get validated payload from context
	validationErrMsg := c.Get("validationErrors")
	if validationErrMsg != "" {
		return c.String(http.StatusBadRequest, validationErrMsg.(string))
	}
	p := c.Get("validatedPayload")
	if p == nil {
		return fmt.Errorf("No validated payload found in the context")
	}
	payload := p.(map[string]interface{})

	// create the entity
	entityType := strings.Split(c.Path(), "/")[1]
	entity, err := models.NewEntity(entityType)
	if err != nil {
		return err
	}
	err = models.PopulateEntity(entity, payload)
	if err != nil {
		return err
	}
	uuid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return err
	}

	// save the entity to database
	db := database.GormDB()
	entity.BaseModel().ID = uuid
	if err = models.UpdateEntity(db, entity); err != nil {
		return err
	}

	return c.String(http.StatusCreated, fmt.Sprintf("%s", entity.BaseModel().ID))
}

// ModelDeleteHandler for request to delete a model entity
func ModelDeleteHandler(c echo.Context) error {
	entityType := strings.Split(c.Path(), "/")[1]
	entity, err := models.NewEntity(entityType)
	if err != nil {
		return err
	}
	uuid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return err
	}
	db := database.GormDB()
	if err = db.Delete(entity, uuid).Error; err != nil {
		return err
	}
	return c.String(http.StatusOK, fmt.Sprintf("%s", uuid.String()))
}
