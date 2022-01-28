package route

import (
	"github.com/jnpr-tjiang/echo-apisvr/pkg/api/handler"
	"github.com/labstack/echo"
)

// AddRPCRoutes add all sync and async RPC routes
func AddRPCRoutes(e *echo.Echo) {
	handler.InitAsyncRPCHandlers()
	e.POST("/rpc/configure", handler.ConfigureRPC)
}
