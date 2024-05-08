package main

import (
	"backend/middleware"
	"backend/routes"

	"github.com/gin-gonic/gin"
	// "github.com/rs/cors"
	"github.com/gin-contrib/cors"
)

func main() {
	// Routers
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Middleware
	router.Use(middleware.CookieAuth())

	// CORS settings
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true //for cookies
	corsConfig.AllowOrigins = []string{"http://localhost:5173", "http://localhost:5173/", "http://127.0.0.1:5173", "http://127.0.0.1:5173/", "https://www.trietandfriends.site/", "https://www.trietandfriends.site"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "access_token"}

	router.Use(cors.New(corsConfig))

	routes.MainRoute(router)
	routes.TaskRoute(router)
	routes.ProjectRoute(router)
	routes.EpicRoute(router)
	routes.AuthRoute(router)
	routes.AccountRoute(router)
	routes.UserInforRoute(router)
	routes.EmployeeRoute(router)
	routes.MessageRoute(router)

	// Old CORS settings
	// corsOptions := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"http://127.0.0.1:5173", "http://localhost:5173", "https://www.trietandfriends.site"},
	// 	AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch, http.MethodPut, http.MethodOptions},
	// 	AllowedHeaders:   []string{"*"},
	// 	AllowCredentials: true,
	// })
	// handler := corsOptions.Handler(router)

	// Server
	router.Run()
}
