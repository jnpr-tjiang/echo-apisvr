package api

import (
	"github.com/jnpr-tjiang/echo-apisvr/pkg/api/route"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/database"
	m "github.com/jnpr-tjiang/echo-apisvr/pkg/middleware"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/models"
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

	// init the data model
	if err := models.Init(); err != nil {
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

	// Start server
	e.Logger.Fatal(e.Start(":7920"))
}
