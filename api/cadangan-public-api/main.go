package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Dev4w4n/e-masjid.my/api/cadangan-public-api/controller"
	_ "github.com/Dev4w4n/e-masjid.my/api/cadangan-public-api/docs"
	"github.com/Dev4w4n/e-masjid.my/api/cadangan-public-api/helper"
	"github.com/Dev4w4n/e-masjid.my/api/cadangan-public-api/repository"
	"github.com/Dev4w4n/e-masjid.my/api/cadangan-public-api/router"
	"github.com/Dev4w4n/e-masjid.my/api/core/env"
	emasjidsaas "github.com/Dev4w4n/e-masjid.my/saas/saas"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	sgin "github.com/go-saas/saas/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title			Cadangan Public Service API
// @version		1.0
// @description	A Cadangan Public  service API in Go using Gin framework
func main() {
	log.Println("Starting server ...")

	env, err := env.GetEnvironment()
	if err != nil {
		log.Fatalf("Error getting environment: %v", err)
	}

	cadanganRepository := repository.NewCadanganRepository()
	cadanganController := controller.NewCadanganController(cadanganRepository)

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{env.AllowOrigins}
	config.AllowMethods = []string{"POST"}

	// Router
	gin.SetMode(gin.ReleaseMode)

	sharedDsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		env.DbHost, env.DbUser, env.DbPassword, env.DbName, env.DbPort)

	emasjidsaas.InitSaas(sharedDsn)

	_router := gin.Default()
	_router.Use(cors.New(config))
	_router.Use(controllerMiddleware())
	_router.Use(sgin.MultiTenancy(emasjidsaas.TenantStorage))

	// enable swagger for dev env
	isLocalEnv := os.Getenv("GO_ENV")
	if isLocalEnv == "local" || isLocalEnv == "dev" {
		_router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	var routes *gin.Engine = _router
	routes = router.NewCadanganPublicRouter(cadanganController, routes, env)

	server := &http.Server{
		Addr:    ":" + env.ServerPort,
		Handler: routes,
	}

	log.Println("Server listening on port ", env.ServerPort)

	err = server.ListenAndServe()
	helper.ErrorPanic(err)
}

// Strictly allow from allowedOrigin
func controllerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the request same-origin is allowed
		secFetchSite := c.Request.Header.Get("Sec-Fetch-Site")

		log.Println("secFetchSite: ", secFetchSite)

		if secFetchSite != "same-origin" && secFetchSite != "same-site" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}
	}
}
