package main

import (
	"fmt"
	"github.com/dotneet/codeapi/storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/dotneet/codeapi/handler"
	"github.com/dotneet/codeapi/template"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	bucket := storage.ImageBucket{
		Endpoint:   viper.GetString("MINIO_ENDPOINT"),
		BucketName: viper.GetString("MINIO_BUCKET_NAME"),
		Secret:     viper.GetString("MINIO_SECRET_KEY"),
		AccessKey:  viper.GetString("MINIO_ACCESS_KEY"),
	}
	handlers := handler.NewHandlers(bucket)

	// Routes
	e.POST("/api/run", handlers.Run)

	// Get port number from viper configuration
	port := viper.GetString("port")
	apibase := viper.GetString("apibase")

	// Serve static files
	// e.Static("/", "public")
	replacer := &template.VariableReplacer{APIBase: apibase}
	e.Group("", replacer.Middleware).Static("/", "public")

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
}

func init() {
	viper.SetConfigType("env")
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	pflag.String("port", "8080", "Set the port number to listen on")
	pflag.String("apibase", "http://localhost:8080", "Set the port number to listen on")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}
