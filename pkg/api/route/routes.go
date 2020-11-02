package route

import (
	"github.com/jnpr-tjiang/echo-apisvr/pkg/api/handler"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/utils"
	"github.com/labstack/echo"
)

// AddCRUDRoutes add all the routes for CRUD operations
func AddCRUDRoutes(e *echo.Echo) error {
	for _, name := range models.ModelNames() {
		path := "/" + utils.Pluralize(name)
		e.POST(path, handler.ModelCreateHandler)
		e.GET(path, handler.ModelGetAllHandler)
		e.GET(path+"/:uuid", handler.ModelGetHandler)
		e.PUT(path+"/:uuid", handler.ModelUpdateHandler)
		e.DELETE(path+"/:uuid", handler.ModelDeleteHandler)
	}
}
