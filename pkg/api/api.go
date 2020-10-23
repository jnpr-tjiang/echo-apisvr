package api

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// Run - run the api server
func Run() {
	e := echo.New()

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes

	// Start server
}
