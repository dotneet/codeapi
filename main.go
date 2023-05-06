package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/dotneet/codeapi/handler"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/api/run", handler.Run)

	// Get port number from viper configuration
	port := viper.GetString("port")

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
}

func init() {
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	pflag.String("port", "8080", "Set the port number to listen on")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}
