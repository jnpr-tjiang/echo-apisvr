package api

import (
	"encoding/json"
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
	"github.com/jnpr-tjiang/echo-apisvr/pkg/utils"
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

func TestValidation(t *testing.T) {
	e := setupTestcase(t)
	status, domainID := createObj(t, e, "domain", `{"name": "default"`)
	require.Equal(t, http.StatusInternalServerError, status)
	require.Contains(t, domainID, "unexpected end of JSON input")

	status, _ = createObj(t, e, "domain", `{"name": "default"}`)
	require.Equal(t, http.StatusCreated, status)
	status, _ = createObj(t, e, "project", `{"name": "juniper", "fq_name": ["default", "juniper"], "display_name": "Juniper Networks"}`)
	require.Equal(t, http.StatusCreated, status)
	status, deviceID := createObj(t, e, "device", `{
		"name": "srx-1",
		"fq_name": ["default", "juniper", "srx-1"],
		"parent_type": "project",
		"region": "(510)xxx-1943",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.2",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED"
	}`)
	require.Equal(t, http.StatusBadRequest, status)
	require.Contains(t, deviceID, "Does not match pattern")
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
	status, deviceFamilyID := createObj(t, e, "device_family", `{
		"fq_name": ["default", "juniper", "srx"]
	}`)
	require.Equal(t, http.StatusCreated, status)
	status, deviceID := createObj(t, e, "device", `{
		"name": "srx-1",
		"fq_name": ["default", "juniper", "srx-1"],
		"parent_type": "project",
		"region": "(510)386-1943",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.2",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED",
		"device_family_refs": [
			{
				"to": ["default", "juniper", "srx"]
			}
		]
	}`)
	require.Equal(t, http.StatusCreated, status)
	status, result = getObjByID(t, e, "device", deviceID, handler.PayloadCfg{})
	require.Equal(t, http.StatusOK, status)
	want = fmt.Sprintf(`{
			"device": {
				"name": "srx-1",
				"uri": "/device/%s",
				"uuid": "%s",
				"fq_name": [
					"default",
					"juniper",
					"srx-1"
				],
				"region": "(510)386-1943",
				"parent_uri": "/project/%s",
				"dic_op_info": {
					"detected_dic_ip": "10.1.1.2",
					"last_detection_timestamp": 13232233.775
				},
				"parent_type": "project",
				"parent_uuid": "%s",
				"display_name": "srx-1",
				"connection_type": "CSP_INITIATED",
				"device_family_refs": [
					{
						"to": ["default", "juniper", "srx"],
						"uuid": "%s",
						"uri": "/device_family/%s",
						"attr": null
					}
				]	
			}
		}`, deviceID, deviceID, projectID, projectID, deviceFamilyID, deviceFamilyID)
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
	status, result := getAllObjs(t, e, "domain", handler.PayloadCfg{
		ShowDetails: false,
	})
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
	status, result = getAllObjs(t, e, "domain", handler.PayloadCfg{
		ShowDetails: true,
	})
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
	require.JSONEq(t, want, result)
}

type responseBasicPayload struct {
	Total  int `json:"total"`
	Device []struct {
		FqName []string `json:"fq_name"`
		UUID   string   `json:"uuid"`
		URI    string   `json:"uri"`
	} `json:"device"`
}

func getObjsWithMultiIDs(t *testing.T, e *echo.Echo, objType string, objUUIDS []string) (int, string) {
	uri := "/" + objType

	if objUUIDS != nil {
		uri += "?obj_uuids="
		for _, item := range objUUIDS {
			uri += item + ","
		}
		uri = strings.TrimRight(uri, ",")
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
func TestGetMultipleObjectUUIDFetch(t *testing.T) {
	e := setupTestcase(t)

	// domain Create
	_, domainID := createObj(t, e, "domain", `{"name": "default"}`)

	// device CRUD
	_, d1 := createObj(t, e, "device", fmt.Sprintf(`{"name": "d1", "fqname": "d1", "parent_uuid": "%s" }`, domainID))
	_, d2 := createObj(t, e, "device", fmt.Sprintf(`{"name": "d2", "fqname": "d2", "parent_uuid": "%s" }`, domainID))
	createObj(t, e, "device", fmt.Sprintf(`{"name": "d3", "fqname": "d3", "parent_uuid": %s }`, domainID))

	// multiple
	status, results := getObjsWithMultiIDs(t, e, "device", []string{d1, d2})

	require.Equal(t, http.StatusOK, status)
	var response responseBasicPayload
	json.Unmarshal([]byte(results), &response)

	expected := []string{d1, d2}
	totalDevicesExpected := 2
	require.Equal(t, totalDevicesExpected, response.Total)

	for _, d := range response.Device {
		require.True(t, true, utils.IndexOf(expected, d.UUID) != -1)
	}
}

func getObjsWithParentIDs(t *testing.T, e *echo.Echo, objType string, parentIDS []string) (int, string) {
	uri := "/" + objType

	if parentIDS != nil {
		uri += "?parent_id="
		for _, item := range parentIDS {
			uri += item + ","
		}
		uri = strings.TrimRight(uri, ",")
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

func TestGetParentIDFilter(t *testing.T) {
	e := setupTestcase(t)

	// domain Create
	_, domainID1 := createObj(t, e, "domain", `{"name": "domain1"}`)
	_, domainID2 := createObj(t, e, "domain", `{"name": "domain2"}`)
	_, domainID4 := createObj(t, e, "domain", `{"name": "domain4"}`)

	// device CRUD
	_, d1 := createObj(t, e, "device", fmt.Sprintf(`{"name": "d1", "fqname": "d1", "parent_uuid": "%s" }`, domainID1))
	_, d2 := createObj(t, e, "device", fmt.Sprintf(`{"name": "d2", "fqname": "d2", "parent_uuid": "%s" }`, domainID2))
	_, d3 := createObj(t, e, "device", fmt.Sprintf(`{"name": "d3", "fqname": "d3", "parent_uuid": "%s" }`, domainID2))

	// extra device
	createObj(t, e, "device", fmt.Sprintf(`{"name": "d4", "fqname": "d4", "parent_uuid": %s }`, domainID4))

	// multiple
	status, results := getObjsWithParentIDs(t, e, "device", []string{domainID1, domainID2, domainID4})
	require.Equal(t, http.StatusOK, status)
	var response responseBasicPayload
	json.Unmarshal([]byte(results), &response)

	expected := []string{d1, d2, d3}
	totalDevicesExpected := 3
	require.Equal(t, totalDevicesExpected, response.Total)

	for _, d := range response.Device {
		require.True(t, true, utils.IndexOf(expected, d.UUID) != -1)
	}
}

func getObjsWithFQNameStr(t *testing.T, e *echo.Echo, objType string, fqnameStr []string) (int, string) {
	uri := "/" + objType

	if fqnameStr != nil {
		uri += "?fq_name_str="
		for _, item := range fqnameStr {
			uri += item + ","
		}
		uri = strings.TrimRight(uri, ",")
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

func TestFqNameFilter(t *testing.T) {
	e := setupTestcase(t)

	// domain Create
	_, domainID := createObj(t, e, "domain", `{"name": "domain"}`)

	// device CRUD
	_, d1 := createObj(t, e, "device", fmt.Sprintf(`{"name": "d1", "fqname": "d1", "parent_uuid": "%s" }`, domainID))
	_, d2 := createObj(t, e, "device", fmt.Sprintf(`{"name": "d2", "fqname": "d2", "parent_uuid": "%s" }`, domainID))
	createObj(t, e, "device", fmt.Sprintf(`{"name": "d3", "fqname": "d3", "parent_uuid": "%s" }`, domainID))

	// multiple
	status, results := getObjsWithFQNameStr(t, e, "device", []string{"domain:d1", "domain:d2"})
	require.Equal(t, http.StatusOK, status)
	var response responseBasicPayload
	json.Unmarshal([]byte(results), &response)

	expected := []string{d1, d2}
	totalDevicesExpected := 2
	require.Equal(t, totalDevicesExpected, response.Total)

	for _, d := range response.Device {
		require.True(t, true, utils.IndexOf(expected, d.UUID) != -1)
	}
}

func TestFieldSelector(t *testing.T) {
	e := setupTestcase(t)

	createObj(t, e, "domain", `{"name": "default"}`)
	status, projectID := createObj(t, e, "project", `{
		"fq_name": ["default", "juniper"], 
		"display_name": "Juniper Networks"}`)
	require.Equal(t, http.StatusCreated, status)
	status, deviceFamilyID := createObj(t, e, "device_family", `{
		"fq_name": ["default", "juniper", "srx"]
	}`)
	require.Equal(t, http.StatusCreated, status)
	status, deviceID := createObj(t, e, "device", `{
		"fq_name": ["default", "juniper", "srx-1"],
		"parent_type": "project",
		"region": "(510)386-1943",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.2",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED",
		"device_family_refs": [
			{
				"to": ["default", "juniper", "srx"]
			}
		]
	}`)
	require.Equal(t, http.StatusCreated, status)

	// ----------------------------------------get entity by id cases---------------------------------------
	// when strict_fields is set, only the basic fields and fields specified in the field selector
	// (including the ref fields) show up
	status, result := getObjByID(t, e, "device", deviceID, handler.PayloadCfg{
		ShowDetails:  false,
		StrictFields: true,
		Fields:       []string{"device_family_refs", "connection_type"},
	})
	require.Equal(t, http.StatusOK, status)
	want := fmt.Sprintf(`{
			"device": {
				"name": "srx-1",
				"uri": "/device/%s",
				"uuid": "%s",
				"fq_name": [
					"default",
					"juniper",
					"srx-1"
				],
				"parent_uri": "/project/%s",
				"parent_type": "project",
				"parent_uuid": "%s",
				"display_name": "srx-1",
				"connection_type": "CSP_INITIATED",
				"device_family_refs": [
					{
						"to": ["default", "juniper", "srx"],
						"uuid": "%s",
						"uri": "/device_family/%s",
						"attr": null
					}
				]
			}
		}`, deviceID, deviceID, projectID, projectID, deviceFamilyID, deviceFamilyID)
	require.JSONEq(t, want, result)

	// fields selector takes precedence than the exclude_refs=false, which means if
	// fields selector exists and does not include the ref fields, the refs won't show up
	// regardless if exclude_refs is set to false or not
	status, result = getObjByID(t, e, "device", deviceID, handler.PayloadCfg{
		ShowDetails:  true,
		StrictFields: true,
		ShowRefs:     true,
		Fields:       []string{"connection_type"},
	})
	require.Equal(t, http.StatusOK, status)
	want = fmt.Sprintf(`{
			"device": {
				"name": "srx-1",
				"uri": "/device/%s",
				"uuid": "%s",
				"fq_name": [
					"default",
					"juniper",
					"srx-1"
				],
				"parent_uri": "/project/%s",
				"parent_type": "project",
				"parent_uuid": "%s",
				"display_name": "srx-1",
				"connection_type": "CSP_INITIATED"
			}
		}`, deviceID, deviceID, projectID, projectID)
	require.JSONEq(t, want, result)

	// without strict_fields, all fields and refs specified in the fields selector will show up. So
	// without strict_fields set to true, the regular fields specified in the field selector does not
	// have any efferct on the outcome
	status, result = getObjByID(t, e, "device", deviceID, handler.PayloadCfg{
		Fields: []string{"device_family_refs", "connection_type"},
	})
	require.Equal(t, http.StatusOK, status)
	want = fmt.Sprintf(`{
			"device": {
				"name": "srx-1",
				"uri": "/device/%s",
				"uuid": "%s",
				"fq_name": [
					"default",
					"juniper",
					"srx-1"
				],
				"parent_uri": "/project/%s",
				"parent_type": "project",
				"parent_uuid": "%s",
				"display_name": "srx-1",
				"region": "(510)386-1943",
				"dic_op_info": {
					"detected_dic_ip": "10.1.1.2",
					"last_detection_timestamp": 13232233.775
				},
				"connection_type": "CSP_INITIATED",
				"device_family_refs": [
					{
						"to": ["default", "juniper", "srx"],
						"uuid": "%s",
						"uri": "/device_family/%s",
						"attr": null
					}
				]
			}
		}`, deviceID, deviceID, projectID, projectID, deviceFamilyID, deviceFamilyID)
	require.JSONEq(t, want, result)

	// test the case where `exclude_refs` is ignored when fields is not set
	status, result = getObjByID(t, e, "device", deviceID, handler.PayloadCfg{
		ShowRefs: false,
	})
	require.Equal(t, http.StatusOK, status)
	want = fmt.Sprintf(`{
			"device": {
				"name": "srx-1",
				"uri": "/device/%s",
				"uuid": "%s",
				"fq_name": [
					"default",
					"juniper",
					"srx-1"
				],
				"parent_uri": "/project/%s",
				"parent_type": "project",
				"parent_uuid": "%s",
				"display_name": "srx-1",
				"region": "(510)386-1943",
				"dic_op_info": {
					"detected_dic_ip": "10.1.1.2",
					"last_detection_timestamp": 13232233.775
				},
				"connection_type": "CSP_INITIATED",
				"device_family_refs": [
					{
						"to": ["default", "juniper", "srx"],
						"uuid": "%s",
						"uri": "/device_family/%s",
						"attr": null
					}
				]
			}
		}`, deviceID, deviceID, projectID, projectID, deviceFamilyID, deviceFamilyID)
	require.JSONEq(t, want, result)

	// ----------------------------------------get all cases---------------------------------------
	status, result = getAllObjs(t, e, "device", handler.PayloadCfg{
		Fields: []string{"device_family_refs", "connection_type"},
	})
	require.Equal(t, http.StatusOK, status)
	want = fmt.Sprintf(`{
			"total": 1,
			"device": [
				{
					"device_family_refs": [
						{
							"to": [
								"default",
								"juniper",
								"srx"
							],
							"attr": null,
							"uri": "/device_family/%s",
							"uuid": "%s"
						}
					],
					"connection_type": "CSP_INITIATED",
					"fq_name": [
						"default",
						"juniper",
						"srx-1"
					],
					"uuid": "%s",
					"uri": "/device/%s"
				}
			]		
		}`, deviceFamilyID, deviceFamilyID, deviceID, deviceID)
	require.JSONEq(t, want, result)

	status, result = getAllObjs(t, e, "device", handler.PayloadCfg{
		ShowDetails: true,
		Fields:      []string{"device_family_refs", "connection_type"},
	})
	require.Equal(t, http.StatusOK, status)
	want = fmt.Sprintf(`{
			"total": 1,
			"device": [
				{
					"name": "srx-1",
					"uri": "/device/%s",
					"uuid": "%s",
					"fq_name": [
						"default",
						"juniper",
						"srx-1"
					],
					"parent_uri": "/project/%s",
					"parent_type": "project",
					"parent_uuid": "%s",
					"display_name": "srx-1",
					"region": "(510)386-1943",
					"dic_op_info": {
						"detected_dic_ip": "10.1.1.2",
						"last_detection_timestamp": 13232233.775
					},
					"connection_type": "CSP_INITIATED",
					"device_family_refs": [
						{
							"to": ["default", "juniper", "srx"],
							"uuid": "%s",
							"uri": "/device_family/%s",
							"attr": null
						}
					]
				}
			]		
		}`, deviceID, deviceID, projectID, projectID, deviceFamilyID, deviceFamilyID)
	require.JSONEq(t, want, result)
}

func TestShowChildren(t *testing.T) {
	e := setupTestcase(t)

	_, domainID := createObj(t, e, "domain", `{"name": "default", "test": "hello"}`)
	_, projectID := createObj(t, e, "project", `{"name": "juniper", "fq_name": ["default", "juniper"], "display_name": "Juniper Networks"}`)
	_, deviceID := createObj(t, e, "device", `{
		"name": "mx",
		"fq_name": ["default", "mx"],
		"parent_type": "domain",
		"region": "(800)386-1943",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.1",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED"}`)

	status, result := getObjByID(t, e, "domain", domainID, handler.PayloadCfg{
		ShowChildren: true,
	})
	want := fmt.Sprintf(`{
		"domain": {
			"name":"default",
			"display_name":"default",
			"fq_name": ["default"],
			"uri":"/domain/%s",
			"uuid":"%s",
			"test": "hello",
			"projects": [
				{
					"uuid": "%s",
					"uri": "/project/%s",
					"to": ["default", "juniper"]
				}
			],
			"devices": [
				{
					"uuid":  "%s",
					"uri": "/device/%s",
					"to": ["default", "mx"]
				}
			]
		}
	}`, domainID, domainID, projectID, projectID, deviceID, deviceID)
	require.Equal(t, http.StatusOK, status)
	require.JSONEq(t, want, result)

	status, result = getObjByID(t, e, "domain", domainID, handler.PayloadCfg{
		Fields: []string{"project"},
	})
	want = fmt.Sprintf(`{
		"domain": {
			"name":"default",
			"display_name":"default",
			"fq_name": ["default"],
			"uri":"/domain/%s",
			"uuid":"%s",
			"test": "hello",
			"projects": [
				{
					"uuid": "%s",
					"uri": "/project/%s",
					"to": ["default", "juniper"]
				}
			]
		}
	}`, domainID, domainID, projectID, projectID)
	require.Equal(t, http.StatusOK, status)
	require.JSONEq(t, want, result)

	status, result = getObjByID(t, e, "domain", domainID, handler.PayloadCfg{
		Fields:       []string{"project"},
		StrictFields: true,
	})
	want = fmt.Sprintf(`{
		"domain": {
			"name":"default",
			"display_name":"default",
			"fq_name": ["default"],
			"uri":"/domain/%s",
			"uuid":"%s",
			"projects": [
				{
					"uuid": "%s",
					"uri": "/project/%s",
					"to": ["default", "juniper"]
				}
			]
		}
	}`, domainID, domainID, projectID, projectID)
	require.Equal(t, http.StatusOK, status)
	require.JSONEq(t, want, result)
}

func TestMultiParentTypes(t *testing.T) {
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

func TestRefCreation(t *testing.T) {
	e := setupTestcase(t)

	// domain CRUD
	createObj(t, e, "domain", `{"name": "default"}`)
	_, projectID := createObj(t, e, "project", `{"name": "juniper", "fq_name": ["default", "juniper"], "display_name": "Juniper Networks"}`)

	// device family
	deviceFamilyID := uuid.New().String()
	status, _ := createObj(t, e, "device_family", fmt.Sprintf(`{
		"uuid": "%s",
		"name": "mx",
		"fq_name": ["default", "juniper", "mx"]}`, deviceFamilyID))
	require.Equal(t, http.StatusCreated, status)

	// device with ref
	deviceID := uuid.New().String()
	status, _ = createObj(t, e, "device", fmt.Sprintf(`{
		"uuid": "%s",
		"name": "mx-1",
		"fq_name": ["default", "juniper", "mx-1"],
		"parent_type": "project",
		"region": "(800)331-5527",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.2",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED",
		"device_family_refs": [
			{
				"uuid": "%s",
				"attr": {
					"test": "foo"
				}
			}
		]}`, deviceID, deviceFamilyID))
	require.Equal(t, http.StatusCreated, status)

	// get device with refs
	var result string
	status, result = getObjByID(t, e, "device", deviceID, handler.PayloadCfg{ShowRefs: true})
	require.Equal(t, http.StatusOK, status)
	want := fmt.Sprintf(`{
		"device": {
			"name": "mx-1",
			"uri": "/device/%s",
			"uuid": "%s",
			"fq_name": ["default", "juniper", "mx-1"],
			"region": "(800)331-5527",
			"parent_uri": "/project/%s",
			"dic_op_info": {
				"detected_dic_ip": "10.1.1.2",
				"last_detection_timestamp": 13232233.775
			},
			"parent_type": "project",
			"parent_uuid": "%s",
			"display_name": "mx-1",
			"connection_type": "CSP_INITIATED",
			"device_family_refs":[
				{
					"uuid": "%s",
					"to": ["default", "juniper", "mx"],
					"uri": "/device_family/%s",
					"attr": {
						"test": "foo"
					}
				}
			]
		}	
	}`, deviceID, deviceID, projectID, projectID, deviceFamilyID, deviceFamilyID)
	require.JSONEq(t, want, result)

	// get device_family with back refs
	status, result = getObjByID(t, e, "device_family", deviceFamilyID, handler.PayloadCfg{ShowBackRefs: true})
	require.Equal(t, http.StatusOK, status)
	want = fmt.Sprintf(`{
		"device_family": {
			"name": "mx",
			"uri": "/device_family/%s",
			"uuid": "%s",
			"fq_name": ["default", "juniper", "mx"],
			"parent_uri": "/project/%s",
			"parent_type": "project",
			"parent_uuid": "%s",
			"display_name": "mx",
			"device_back_refs":[
				{
					"uuid": "%s",
					"to": ["default", "juniper", "mx-1"],
					"uri": "/device/%s",
					"attr": {
						"test": "foo"
					}
				}
			]
		}	
	}`, deviceFamilyID, deviceFamilyID, projectID, projectID, deviceID, deviceID)
	require.JSONEq(t, want, result)
}

func TestPartialUpdate(t *testing.T) {
	e := setupTestcase(t)

	// domain CRUD
	createObj(t, e, "domain", `{"name": "default"}`)
	_, projectID := createObj(t, e, "project", `{"name": "juniper", "fq_name": ["default", "juniper"], "display_name": "Juniper Networks"}`)

	// device with ref
	status, deviceID := createObj(t, e, "device", fmt.Sprintf(`{
		"name": "mx-1",
		"fq_name": ["default", "juniper", "mx-1"],
		"parent_type": "project",
		"region": "(800)331-5527",
		"dic_op_info": {
			"detected_dic_ip": "10.1.1.2",
			"last_detection_timestamp": 13232233.775
		},
		"connection_type": "CSP_INITIATED"
	}`))
	require.Equal(t, http.StatusCreated, status)

	status, _ = updateObj(t, e, "device", deviceID, fmt.Sprintf(`{
		"connection_type": "DEVICE_INITIATED"
	}`))
	require.Equal(t, http.StatusOK, status)

	status, result := getObjByID(t, e, "device", deviceID, handler.PayloadCfg{})
	want := fmt.Sprintf(`{
		"device": {
			"name": "mx-1",
			"uri": "/device/%s",
			"uuid": "%s",
			"fq_name": ["default", "juniper", "mx-1"],
			"region": "(800)331-5527",
			"parent_uri": "/project/%s",
			"dic_op_info": {
				"detected_dic_ip": "10.1.1.2",
				"last_detection_timestamp": 13232233.775
			},
			"parent_type": "project",
			"parent_uuid": "%s",
			"display_name": "mx-1",
			"connection_type": "DEVICE_INITIATED"
		}	
	}`, deviceID, deviceID, projectID, projectID)
	require.JSONEq(t, want, result)

	status, _ = updateObj(t, e, "device", deviceID, fmt.Sprintf(`{
		"region": "(800)386-1943",
		"connection_type": "DEVICE_INITIATED"
	}`))
	require.Equal(t, http.StatusOK, status)

	status, result = getObjByID(t, e, "device", deviceID, handler.PayloadCfg{})
	want = fmt.Sprintf(`{
		"device": {
			"name": "mx-1",
			"uri": "/device/%s",
			"uuid": "%s",
			"fq_name": ["default", "juniper", "mx-1"],
			"region": "(800)386-1943",
			"parent_uri": "/project/%s",
			"dic_op_info": {
				"detected_dic_ip": "10.1.1.2",
				"last_detection_timestamp": 13232233.775
			},
			"parent_type": "project",
			"parent_uuid": "%s",
			"display_name": "mx-1",
			"connection_type": "DEVICE_INITIATED"
		}	
	}`, deviceID, deviceID, projectID, projectID)
	require.JSONEq(t, want, result)
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
	if rec.Code != http.StatusCreated {
		return rec.Code, rec.Body.String()
	}
	domainID, err = uuid.Parse(rec.Body.String())
	require.NoError(t, err)
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

func updateObj(t *testing.T, e *echo.Echo, objType string, objID string, payload string) (int, string) {
	rec := executeRequest(t, e, RequestInfo{
		method:         http.MethodPut,
		uri:            fmt.Sprintf("/%s/%s", objType, objID),
		payload:        payload,
		middlewareFunc: middleware.JSONSchemaValidator(),
		handlerFunc:    handler.ModelUpdateHandler,
		ctxInit: func(c echo.Context) {
			c.SetPath(fmt.Sprintf("/%s/:id", objType))
			c.SetParamNames("id")
			c.SetParamValues(objID)
		},
	})
	return http.StatusOK, rec.Body.String()
}

func getAllObjs(t *testing.T, e *echo.Echo, objType string, cfg handler.PayloadCfg) (int, string) {
	rec := executeRequest(t, e, RequestInfo{
		method:         http.MethodGet,
		uri:            fmt.Sprintf("/%s%s", objType, toQueryStr(cfg)),
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
	(*cfg).Server.Schema = os.Getenv("SCHEMA")

	e := echo.New()
	e.Debug = true
	_, err := database.Init(cfg)
	require.NoError(t, err)
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
