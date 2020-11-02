package api

import (
	"fmt"
	"net/http"

	"github.com/jnpr-tjiang/echo-apisvr/pkg/database"
	m "github.com/jnpr-tjiang/echo-apisvr/pkg/middleware"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func loadRewriteConfig() middleware.RewriteConfig {
	rules := make(map[string]string)
	rules["/domains"] = "/_crud?type=domain"
	rules["/projects"] = "/_crud?type=project"
	rules["/devices"] = "/_crud?type=device"
	rules["/devicefamilies"] = "/crud?type=devicefamily"
	return middleware.RewriteConfig{
		Rules: rules,
	}
}

// Run - run the api server
func Run() {
	e := echo.New()
	e.Debug = true

	// initialize database
	if _, err := database.Init(); err != nil {
		panic(err)
	}

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// e.Pre(middleware.RewriteWithConfig(loadRewriteConfig()))
	stats := m.NewAPIStats()
	e.Use(stats.Middleware())
	e.GET("/stats", stats.GetStatsHandler())

	// Routes
	// routes.AddCRUDRoutes()
	e.GET("/domains/:id", getCreateHandler("domain"))

	// Start server
	e.Logger.Fatal(e.Start(":7920"))
}

func getCreateHandler(objType string) echo.HandlerFunc {
	return func(c echo.Context) error {
		msg := fmt.Sprintf("url: %s\ntype: %s\nid: %s\nqstr=%s\n", c.Request().URL, objType, c.Param("id"), c.QueryParam("qstr"))
		return c.String(http.StatusOK, msg)
	}
}
