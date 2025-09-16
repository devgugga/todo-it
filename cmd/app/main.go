package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devgugga/todo-it/internal/config"
	"github.com/devgugga/todo-it/internal/database"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.LoadConfig()

	mongoConfig := &database.MongoConfig{
		URI:            cfg.MongoURI,
		DBName:         cfg.MongoDBName,
		MaxPoolSize:    20,
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
	}

	db, err := database.NewMongoClient(mongoConfig)
	if err != nil {
		log.Fatalf("❌ Falha ao conectar com MongoDB: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("❌ Erro ao fechar conexão MongoDB: %v", err)
		}
	}()

	app := fiber.New(fiber.Config{
		AppName:      "Todo API v1.0",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			log.Printf("❌ Erro na API: %v", err)
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
				"code":    code,
			})
		},
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("🔄 Iniciando graceful shutdown...")

		if err := app.Shutdown(); err != nil {
			log.Printf("❌ Erro durante shutdown: %v", err)
		}
	}()

	log.Printf("🚀 Servidor rodando na porta %s", cfg.Port)
	log.Printf("📊 Health check: http://localhost:%s/health", cfg.Port)
	log.Printf("📚 API Base: http://localhost:%s/api/v1", cfg.Port)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
	}
}
