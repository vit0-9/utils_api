// @title           Utility API
// @version         1.0
// @description     A collection of useful utilities including network, URL, and web analysis tools.

// @contact.name   API Support
// @contact.email  info@bentech.app

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
// @schemes   http https
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/vit0-9/utils_api/docs"   // Your Swagger docs
	"github.com/vit0-9/utils_api/handlers" // Your handlers package
)

// App encapsulates all the components of the application
type App struct {
	Router              *gin.Engine
	NetIntelHandlers    *handlers.NetworkIntelligenceHandlers
	URLUtilHandlers     *handlers.URLUtilitiesHandlers
	WebAnalysisHandlers *handlers.WebAnalysisHandlers
	HealthHandler       *handlers.HealthHandler
}

// NewApp creates and initializes a new application instance
func NewApp() (*App, error) {
	netIntelHandlers := handlers.NewNetworkIntelligenceHandlers()
	urlUtilHandlers := handlers.NewURLUtilitiesHandlers()
	webAnalysisHandlers := handlers.NewWebAnalysisHandlers()
	healthHandler := handlers.NewHealthHandler()

	router := gin.Default()
	// Consider your proxy setup for SetTrustedProxies if deploying
	// err := router.SetTrustedProxies(nil)
	// if err != nil {
	// 	log.Printf("Warning: Could not set trusted proxies: %v", err)
	// }

	app := &App{
		Router:              router,
		NetIntelHandlers:    netIntelHandlers,
		URLUtilHandlers:     urlUtilHandlers,
		WebAnalysisHandlers: webAnalysisHandlers,
		HealthHandler:       healthHandler,
	}

	app.setupRoutes()
	return app, nil
}

// setupRoutes defines all the application routes
func (app *App) setupRoutes() {
	// Health check endpoint (can be top-level)
	// For Swagger, this will be documented relative to @host if its @Router path starts with /
	app.Router.GET("/api/v1/health", app.HealthHandler.HealthCheckHandler)

	// Group for Network & Domain Intelligence utilities
	// These will be prefixed by @BasePath /api/v1
	netIntelV1 := app.Router.Group("/api/v1/net")
	{
		netIntelV1.GET("/dns-lookup", app.NetIntelHandlers.DNSLookupHandler)
		netIntelV1.GET("/ip-info", app.NetIntelHandlers.IPInfoHandler)
		netIntelV1.GET("/whois-lookup", app.NetIntelHandlers.WhoisLookupHandler)
		netIntelV1.GET("/ssl-check", app.NetIntelHandlers.SSLCheckHandler)
	}

	// Group for URL Manipulation utilities
	urlUtilV1 := app.Router.Group("/api/v1/url")
	{
		urlUtilV1.POST("/clean", app.URLUtilHandlers.CleanURLHandler)
		urlUtilV1.GET("/resolve-redirect", app.URLUtilHandlers.ResolveRedirectHandler)
		urlUtilV1.POST("/generate-utm", app.URLUtilHandlers.GenerateUTMHandler)
	}

	// Group for Web Analysis utilities
	webAnalysisV1 := app.Router.Group("/api/v1/web")
	{
		webAnalysisV1.GET("/stack-analyzer", app.WebAnalysisHandlers.StackAnalyzerHandler)
		webAnalysisV1.GET("/http-headers", app.WebAnalysisHandlers.HTTPHeadersHandler)
	}

	// Add Swagger route
	// This path should be absolute from the host, not affected by @BasePath
	app.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))
	// It's good practice to explicitly set the URL for doc.json for clarity
	// The default might be swagger.json or docs.json depending on swag version/config
}

// Start runs the Gin HTTP server
func (app *App) Start(addr string) error {
	log.Printf("🚀 API server starting on %s", addr)
	return app.Router.Run(addr)
}
