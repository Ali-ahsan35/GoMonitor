package main

import (
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"gomonitor/backend/handler"
	"gomonitor/backend/pkg/httpclient"
	"gomonitor/backend/pkg/profiler"
	"gomonitor/backend/service"
)

func main() {
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	allowedOrigins := resolveAllowedOrigins()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Api-Key"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	pprof.Register(router, "/debug/pprof")

	client := httpclient.New(15 * time.Second)
	testService := service.NewTestService(client)
	testHandler := handler.NewTestHandler(testService)
	profileService := service.NewProfileService(profiler.New())
	profileHandler := handler.NewProfileHandler(profileService)

	router.POST("/run-test", testHandler.RunTest)
	router.GET("/profiles/types", profileHandler.Types)
	router.POST("/profiles/capture", profileHandler.Capture)
	router.GET("/profiles/:type/download", profileHandler.Download)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("GoMonitor backend listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func resolveAllowedOrigins() []string {
	originEnv := strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN"))
	if originEnv == "" {
		return []string{"http://localhost:5173"}
	}

	parts := strings.Split(originEnv, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return []string{"http://localhost:5173"}
	}

	return origins
}
