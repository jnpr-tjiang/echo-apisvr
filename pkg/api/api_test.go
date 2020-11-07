package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/api/handler"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/api/route"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/config"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/database"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/middleware"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
}

func shutdown() {

}

func TestCRUD(t *testing.T) {
	e := setupTestcase(t)

	// domain CRUD
	domainID := createObj(t, e, "domain", `{"name": "default"}`)
	result := getObjByID(t, e, "domain", domainID)
	want := fmt.Sprintf(`{
		"domain": {
			"name":"default",
			"display_name":"default",
			"fq_name": ["default"],
			"uri":"/domain/%s",
			"uuid":"%s"
		}
	}`, domainID, domainID)
	require.JSONEq(t, want, result)

	// project CRUD
	projectID := createObj(t, e, "project", `{"name": "juniper", "fq_name": ["default", "juniper"], "display_name": "Juniper Networks"}`)
	result = getObjByID(t, e, "project", projectID)
	want = fmt.Sprintf(`{
			"project": {
				"name":"juniper",
				"display_name":"Juniper Networks",
				"fq_name": ["default", "juniper"],
				"uri":"/project/%s",
				"uuid":"%s",
				"parent_type": "domain",
				"parent_uuid": "%s",
				"parent_uri": "/domain/%s"
			}
		}`, projectID, projectID, domainID, domainID)
	require.JSONEq(t, want, result)

	// device
	deviceID := createObj(t, e, "device", `{
		"name": "junos",
		"fq_name": ["default", "juniper", "junos"],
		"region": "(510)386-1943",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.2",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED"
	}`)
	result = getObjByID(t, e, "device", deviceID)
	want = fmt.Sprintf(`{
			"device": {
				"name": "junos",
				"uri": "/device/%s",
				"uuid": "%s",
				"fq_name": [
					"default",
					"juniper",
					"junos"
				],
				"region": "(510)386-1943",
				"parent_uri": "/project/%s",
				"dic_op_info": {
					"detected_dic_ip": "10.1.1.2",
					"last_detection_timestamp": 13232233.775
				},
				"parent_type": "project",
				"parent_uuid": "%s",
				"display_name": "junos",
				"connection_type": "CSP_INITIATED"
			}
		}`, deviceID, deviceID, projectID, projectID)
	require.JSONEq(t, want, result)

	// deletion
	result = deleteObj(t, e, "device", deviceID)
	require.Equal(t, deviceID, result)
	result = deleteObj(t, e, "project", projectID)
	require.Equal(t, projectID, result)
	result = deleteObj(t, e, "domain", domainID)
	require.Equal(t, domainID, result)
}

type RequestInfo struct {
	method         string
	uri            string
	payload        string
	middlewareFunc echo.MiddlewareFunc
	handlerFunc    echo.HandlerFunc
	ctxInit        func(c echo.Context)
}

func createObj(t *testing.T, e *echo.Echo, objType string, payload string) string {
	rec := executeRequest(t, e, RequestInfo{
		method:         http.MethodPost,
		uri:            "/" + objType,
		payload:        payload,
		middlewareFunc: middleware.JSONSchemaValidator(),
		handlerFunc:    handler.ModelCreateHandler,
		ctxInit: func(c echo.Context) {
			c.SetPath("/" + objType)
		},
	})
	require.Equal(t, http.StatusCreated, rec.Code)
	domainID, err := uuid.Parse(rec.Body.String())
	require.NoError(t, err)
	return domainID.String()
}

func getObjByID(t *testing.T, e *echo.Echo, objType string, objID string) string {
	rec := executeRequest(t, e, RequestInfo{
		method:         http.MethodGet,
		uri:            fmt.Sprintf("/%s/%s", objType, objID),
		payload:        "",
		middlewareFunc: nil,
		handlerFunc:    handler.ModelGetHandler,
		ctxInit: func(c echo.Context) {
			c.SetPath(fmt.Sprintf("/%s/:id", objType))
			c.SetParamNames("id")
			c.SetParamValues(objID)
		},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	return rec.Body.String()
}

func deleteObj(t *testing.T, e *echo.Echo, objType string, objID string) string {
	rec := executeRequest(t, e, RequestInfo{
		method:         http.MethodDelete,
		uri:            fmt.Sprintf("/%s/%s", objType, objID),
		payload:        "",
		middlewareFunc: nil,
		handlerFunc:    handler.ModelDeleteHandler,
		ctxInit: func(c echo.Context) {
			c.SetPath(fmt.Sprintf("/%s/:id", objType))
			c.SetParamNames("id")
			c.SetParamValues(objID)
		},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	return rec.Body.String()
}

func setupTestcase(t *testing.T) *echo.Echo {
	os.Remove("test.db")
	config.InitConfig("")
	cfg := config.GetConfig()
	(*cfg).Database.Driver = "sqlite3"
	(*cfg).Database.Dbname = "test"

	e := echo.New()
	e.Debug = true
	_, err := database.Init(cfg)
	require.NoError(t, err)
	require.NoError(t, models.Init())
	route.AddCRUDRoutes(e)
	return e
}

func executeRequest(t *testing.T, e *echo.Echo, info RequestInfo) *httptest.ResponseRecorder {
	// create the request
	req := httptest.NewRequest(info.method, info.uri, strings.NewReader(info.payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	info.ctxInit(c)

	// run the JSONSchemaValidator middleware
	if info.middlewareFunc != nil {
		handle := info.middlewareFunc(echo.HandlerFunc(func(c echo.Context) error {
			return nil
		}))
		handle(c)
	}

	info.handlerFunc(c)
	if rec.Code == http.StatusBadRequest {
		t.Logf("Invalid request payload: %s", rec.Body.String())
	}
	return rec
}
