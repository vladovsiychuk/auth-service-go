package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vladovsiychuk/auth-service-go/configs"
	"github.com/vladovsiychuk/auth-service-go/internal/auth"
	"github.com/vladovsiychuk/auth-service-go/pkg/helper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	r := gin.Default()
	r.Use(CORSMiddleware())

	postgresDB := setupPostgres()
	configs.SetupDbMigration(postgresDB)

	injectDependencies(postgresDB, r)

	r.Run(":8080")
}

func injectDependencies(
	postgresDB *gorm.DB,
	r *gin.Engine,
) {
	sessionTokenRepository := auth.NewSessionTokenRepository(postgresDB)

	authRepository := auth.NewKeyRepository(postgresDB)
	authService := auth.NewService(authRepository, sessionTokenRepository)
	authService.Init()
	authHandler := auth.NewRouter(authService)
	authHandler.RegisterRoutes(r)
}

func setupPostgres() *gorm.DB {
	host := helper.GetEnv("POSTGRES_HOST", "localhost")
	user := helper.GetEnv("POSTGRES_USER", "root")
	password := helper.GetEnv("POSTGRES_PASSWORD", "rootpassword")
	dbname := helper.GetEnv("POSTGRES_DB_NAME", "postgres")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", host, user, password, dbname)
	postgresDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return postgresDB
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
