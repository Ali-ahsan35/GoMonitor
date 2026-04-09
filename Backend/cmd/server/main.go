package main

import (
	"log"
	"runtime"
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

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
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

	log.Println("GoMonitor backend listening on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
