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
)

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
	entity, err := newEntity(c.Path())
	if err != nil {
		return err
	}
	populateBaseModel(entity.BaseModel(), payload)

	// save the entity to database
	db := database.GormDB()
	if err = db.Create(entity).Error; err != nil {
		return err
	}
	return c.String(http.StatusOK, fmt.Sprintf("%s", entity.BaseModel().ID))
}

func newEntity(path string) (models.Entity, error) {
	entityType := strings.Split(path, "/")[1]
	entity, err := models.NewEntity(utils.Singularize(entityType))
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// ModelGetAllHandler for request to get all model entities
func ModelGetAllHandler(c echo.Context) error {
	msg := fmt.Sprintf("url: %s\nid: %s\nqstr=%s\n", c.Request().URL, c.Param("id"), c.QueryParam("qstr"))
	return c.String(http.StatusOK, msg)
}

// ModelGetHandler for request to get an model entity by id
func ModelGetHandler(c echo.Context) error {
	entity, err := newEntity(c.Path())
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
	return c.String(http.StatusOK, fmt.Sprintf("%s", entity.BaseModel().Payload))
}

// ModelUpdateHandler for request to update a model entity
func ModelUpdateHandler(c echo.Context) error {
	msg := fmt.Sprintf("url: %s\nid: %s\nqstr=%s\n", c.Request().URL, c.Param("id"), c.QueryParam("qstr"))
	return c.String(http.StatusOK, msg)
}

// ModelDeleteHandler for request to delete a model entity
func ModelDeleteHandler(c echo.Context) error {
	msg := fmt.Sprintf("url: %s\nid: %s\nqstr=%s\n", c.Request().URL, c.Param("id"), c.QueryParam("qstr"))
	return c.String(http.StatusOK, msg)
}

func populateBaseModel(m *models.BaseModel, payload map[string]interface{}) {
	if ID, ok := payload["uuid"]; ok {
		m.ID = ID.(uuid.UUID)
	}
	if name, ok := payload["name"]; ok {
		m.Name = name.(string)
	}
	if fqname, ok := payload["fqname"]; ok {
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
	m.Payload, _ = json.Marshal(payload)
}