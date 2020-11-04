package route

import (
	"github.com/jnpr-tjiang/echo-apisvr/pkg/api/handler"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/middleware"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models"
	"github.com/labstack/echo"
)

// AddCRUDRoutes add all the routes for CRUD operations
func AddCRUDRoutes(e *echo.Echo) {
	for _, name := range models.ModelNames() {
		path := "/" + name
		e.POST(path, handler.ModelCreateHandler, middleware.JSONSchemaValidator())
		e.GET(path+"/:id", handler.ModelGetHandler)
		e.PUT(path+"/:id", handler.ModelUpdateHandler, middleware.JSONSchemaValidator())
		e.DELETE(path+"/:id", handler.ModelDeleteHandler)
	}
}
