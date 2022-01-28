package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/jnpr-tjiang/echo-apisvr/pkg/config"
	"github.com/labstack/echo"
	"github.com/xeipuuv/gojsonschema"
)

var (
	schema *gojsonschema.Schema
	err    error
)

// JSONSchemaValidator middleware validate the payload against the model schema
func JSONSchemaValidator() echo.MiddlewareFunc {
	if schema == nil {
		schemaFile := config.GetConfig().Server.Schema
		sl := gojsonschema.NewReferenceLoader(fmt.Sprintf("file:///%s", schemaFile))
		schema, err = gojsonschema.NewSchema(sl)
		if err != nil {
			panic(err)
		}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var validationErrMsg string
			buf := new(bytes.Buffer)
			len, err := buf.ReadFrom(c.Request().Body)
			if err != nil {
				return err
			}
			if len > 0 {
				var payload interface{}
				if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
					c.Error(err)
					c.Set("validationErrors", err.Error())
					return err
				}
				dl := gojsonschema.NewGoLoader(payload)
				result, err := schema.Validate(dl)
				if err != nil {
					c.Error(err)
					c.Set("validationErrors", err.Error())
					return err
				}
				for _, verr := range result.Errors() {
					validationErrMsg += verr.String() + "\n"
				}
				c.Set("validationErrors", validationErrMsg)
				c.Set("validatedPayload", payload)
				if err := next(c); err != nil {
					c.Error(err)
				}
			}
			return nil
		}
	}
}
