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

func TestBasicCRUD(t *testing.T) {
	e := setupTestcase(t)

	// domain CRUD
	status, domainID := createObj(t, e, "domain", `{"name": "default"}`)
	require.Equal(t, http.StatusCreated, status)

	status, result := getObjByID(t, e, "domain", domainID, handler.PayloadCfg{})
	require.Equal(t, http.StatusOK, status)
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
	status, projectID := createObj(t, e, "project", `{"name": "juniper", "fq_name": ["default", "juniper"], "display_name": "Juniper Networks"}`)
	require.Equal(t, http.StatusCreated, status)

	status, result = getObjByID(t, e, "project", projectID, handler.PayloadCfg{})
	require.Equal(t, http.StatusOK, status)
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
	status, deviceID := createObj(t, e, "device", `{
		"name": "junos",
		"fq_name": ["default", "juniper", "junos"],
		"region": "(510)386-1943",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.2",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED"
	}`)
	require.Equal(t, http.StatusCreated, status)
	status, result = getObjByID(t, e, "device", deviceID, handler.PayloadCfg{})
	require.Equal(t, http.StatusOK, status)
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
	status, result = deleteObj(t, e, "device", deviceID)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, deviceID, result)
	status, result = deleteObj(t, e, "project", projectID)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, projectID, result)
	status, result = deleteObj(t, e, "domain", domainID)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, domainID, result)
}

func TestGetAll(t *testing.T) {
	e := setupTestcase(t)

	// domain CRUD
	_, id1 := createObj(t, e, "domain", `{"name": "domain_1"}`)
	_, id2 := createObj(t, e, "domain", `{"name": "domain_2"}`)

	// default
	status, result := getAllObjs(t, e, "domain", false)
	require.Equal(t, http.StatusOK, status)
	want := fmt.Sprintf(`{
		"total": 2,
		"domain": [
			{
				"fq_name": ["domain_1"],
				"uri":"/domain/%s",
				"uuid":"%s"
			},
			{
				"fq_name": ["domain_2"],
				"uri":"/domain/%s",
				"uuid":"%s"
			}
		]
	}`, id1, id1, id2, id2)
	require.JSONEq(t, want, result)

	// detail=true
	status, result = getAllObjs(t, e, "domain", true)
	require.Equal(t, http.StatusOK, status)
	want = fmt.Sprintf(`{
		"total": 2,
		"domain": [
			{
				"name": "domain_1",
				"display_name": "domain_1",
				"fq_name": ["domain_1"],
				"uri":"/domain/%s",
				"uuid":"%s"
			},
			{
				"name": "domain_2",
				"display_name": "domain_2",
				"fq_name": ["domain_2"],
				"uri":"/domain/%s",
				"uuid":"%s"
			}
		]
	}`, id1, id1, id2, id2)
	t.Logf("want:\n%s\n", want)
	t.Logf("actual:\n%s\n", result)
	require.JSONEq(t, want, result)
}

func TestFieldFilter(t *testing.T) {
	e := setupTestcase(t)

	createObj(t, e, "domain", `{"name": "default"}`)
	_, projectID := createObj(t, e, "project", `{
		"name": "juniper", 
		"fq_name": ["default", "juniper"], 
		"display_name": "Juniper Networks"}`)
	_, deviceID := createObj(t, e, "device", `{
		"name": "junos",
		"fq_name": ["default", "juniper", "junos"],
		"region": "(510)386-1943",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.2",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED"}`)

	status, result := getObjByID(t, e, "device", deviceID, handler.PayloadCfg{
		StrictFields: true,
		Fields:       []string{"connection_type"},
	})
	require.Equal(t, http.StatusOK, status)
	want := fmt.Sprintf(`{
			"device": {
				"name": "junos",
				"uri": "/device/%s",
				"uuid": "%s",
				"fq_name": [
					"default",
					"juniper",
					"junos"
				],
				"parent_uri": "/project/%s",
				"parent_type": "project",
				"parent_uuid": "%s",
				"display_name": "junos",
				"connection_type": "CSP_INITIATED"
			}
		}`, deviceID, deviceID, projectID, projectID)
	require.JSONEq(t, want, result)
}

func TestMultiParent(t *testing.T) {
	e := setupTestcase(t)

	// domain CRUD
	createObj(t, e, "domain", `{"name": "default"}`)
	createObj(t, e, "project", `{"name": "juniper", "fq_name": ["default", "juniper"], "display_name": "Juniper Networks"}`)

	// device
	status, _ := createObj(t, e, "device", `{
		"name": "mx",
		"fq_name": ["default", "mx"],
		"parent_type": "domain",
		"region": "(800)386-1943",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.1",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED"}`)
	require.Equal(t, http.StatusCreated, status)
	status, _ = createObj(t, e, "device", `{
		"name": "srx",
		"fq_name": ["default", "juniper", "srx"],
		"parent_type": "project",
		"region": "(800)331-5527",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.2",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED"}`)
	require.Equal(t, http.StatusCreated, status)
}

type RequestInfo struct {
	method         string
	uri            string
	payload        string
	middlewareFunc echo.MiddlewareFunc
	handlerFunc    echo.HandlerFunc
	ctxInit        func(c echo.Context)
}

func createObj(t *testing.T, e *echo.Echo, objType string, payload string) (int, string) {
	rec := executeRequest(t, e, RequestInfo{
		method:         http.MethodPost,
		uri:            "/" + objType,
		payload:        payload,
		middlewareFunc: middleware.JSONSchemaValidator(),
		handlerFunc:    handler.ModelCreateHandler,
		ctxInit: func(c echo.Context) {
			c.SetPath(fmt.Sprintf("/%s", objType))
		},
	})
	var (
		domainID uuid.UUID
		err      error
	)
	if rec.Code == http.StatusCreated {
		domainID, err = uuid.Parse(rec.Body.String())
		require.NoError(t, err)
	}
	return rec.Code, domainID.String()
}

func toQueryStr(cfg handler.PayloadCfg) string {
	qstr := "?"
	if cfg.ShowDetails {
		qstr += "detail=true&"
	}
	if cfg.StrictFields {
		qstr += "strict_fields&"
	}
	if len(cfg.Fields) > 0 {
		qstr += "fields="
		for i, field := range cfg.Fields {
			qstr += field
			if (i + 1) < len(cfg.Fields) {
				qstr += ","
			} else {
				qstr += "&"
			}
		}
	}
	if cfg.ShowRefs {
		qstr += "exclude_refs=false&"
	}
	if cfg.ShowBackRefs {
		qstr += "exclude_back_refs=false&"
	}
	if cfg.ShowChildren {
		qstr += "exclude_children=false&"
	}
	return qstr
}

func getObjByID(t *testing.T, e *echo.Echo, objType string, objID string, cfg handler.PayloadCfg) (int, string) {
	rec := executeRequest(t, e, RequestInfo{
		method:         http.MethodGet,
		uri:            fmt.Sprintf("/%s/%s%s", objType, objID, toQueryStr(cfg)),
		payload:        "",
		middlewareFunc: nil,
		handlerFunc:    handler.ModelGetHandler,
		ctxInit: func(c echo.Context) {
			c.SetPath(fmt.Sprintf("/%s/:id", objType))
			c.SetParamNames("id")
			c.SetParamValues(objID)
		},
	})
	return http.StatusOK, rec.Body.String()
}

func getAllObjs(t *testing.T, e *echo.Echo, objType string, detail bool) (int, string) {
	uri := "/" + objType
	if detail {
		uri += "?detail=true"
	}
	rec := executeRequest(t, e, RequestInfo{
		method:         http.MethodGet,
		uri:            uri,
		payload:        "",
		middlewareFunc: nil,
		handlerFunc:    handler.ModelGetAllHandler,
		ctxInit: func(c echo.Context) {
			c.SetPath(fmt.Sprintf("/%s", objType))
		},
	})
	return rec.Code, rec.Body.String()
}

func deleteObj(t *testing.T, e *echo.Echo, objType string, objID string) (int, string) {
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
	return rec.Code, rec.Body.String()
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
