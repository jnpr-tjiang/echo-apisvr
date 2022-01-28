package api

import (
	"github.com/jnpr-tjiang/echo-apisvr/pkg/api/route"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/config"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/database"
	m "github.com/jnpr-tjiang/echo-apisvr/pkg/middleware"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// Run - run the api server
func Run() {
	e := echo.New()
	e.Debug = true

	// initialize database
	if _, err := database.Init(config.GetConfig()); err != nil {
		panic(err)
	}

	// middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	stats := m.NewAPIStats()
	e.Use(stats.Middleware())
	e.GET("/stats", stats.GetStatsHandler())

	// Routes
	route.AddCRUDRoutes(e)
	route.AddRPCRoutes(e)

	// Start server
	e.Logger.Fatal(e.Start(":7920"))
}
