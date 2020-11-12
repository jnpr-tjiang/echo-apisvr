package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/database"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models/custom"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/utils"
	"github.com/labstack/echo"
	"gorm.io/gorm"
)

type PayloadCfg struct {
	ShowDetails  bool
	StrictFields bool
	Fields       []string
	ShowRefs     bool
	ShowBackRefs bool
	ShowChildren bool
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

func buildEntityPayload(db *gorm.DB, entity models.Entity, cfg PayloadCfg) (payload []byte, err error) {
	if !cfg.ShowDetails {
		uuid := entity.BaseModel().ID.String()
		payload := fmt.Sprintf(
			`{"fq_name":%s,"uuid":"%s","uri":"/%s/%s"}`,
			entity.BaseModel().FQName, uuid, strings.ToLower(utils.TypeOf(entity)), uuid)
		return []byte(payload), nil
	}

	var jsonPayload map[string]interface{}
	if cfg.StrictFields && len(cfg.Fields) > 0 {
		if jsonPayload, err = filterFields(entity.BaseModel().Payload, cfg.Fields); err != nil {
			return []byte{}, err
		}
	}
	if cfg.ShowRefs {

	}
	if cfg.ShowBackRefs {

	}
	if cfg.ShowChildren {

	}
	if len(jsonPayload) > 0 {
		return json.Marshal(jsonPayload)
	}
	return entity.BaseModel().Payload, nil
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
	populateBaseModel(entity.BaseModel(), payload)

	// save the entity to database
	db := database.GormDB()
	if err = db.Create(entity).Error; err != nil {
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
	msg := fmt.Sprintf("url: %s\nid: %s\nqstr=%s\n", c.Request().URL, c.Param("id"), c.QueryParam("qstr"))
	return c.String(http.StatusOK, msg)
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

func populateBaseModel(m *models.BaseModel, payload map[string]interface{}) {
	if ID, ok := payload["uuid"]; ok {
		m.ID = ID.(uuid.UUID)
	}
	if name, ok := payload["name"]; ok {
		m.Name = name.(string)
	}
	if displayName, ok := payload["display_name"]; ok {
		m.DisplayName = displayName.(string)
	}
	if fqname, ok := payload["fq_name"]; ok {
		var s []string
		for _, v := range fqname.([]interface{}) {
			s = append(s, v.(string))
		}
		fqn := custom.FQName(s)
		if val, err := custom.FQName(fqn).Value(); err == nil {
			m.FQName = val.(string)
		}
	}
	if parentType, ok := payload["parent_type"]; ok {
		m.ParentType = parentType.(string)
	}
	if parentID, ok := payload["parent_uuid"]; ok {
		m.ParentID = parentID.(uuid.UUID)
	}
	m.JSON = &payload
}
