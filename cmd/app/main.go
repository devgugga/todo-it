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
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	log.Println("🚀 Iniciando Todo API...")

	// Carrega configurações
	cfg := config.LoadConfig()

	// Configura MongoDB
	mongoConfig := &database.MongoConfig{
		URI:            cfg.MongoURI,
		DBName:         cfg.MongoDBName,
		MaxPoolSize:    20,
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
	}

	// Inicializa o banco de dados (cria collections, índices, etc.)
	db, err := database.InitializeDatabase(mongoConfig)
	if err != nil {
		log.Fatalf("❌ Falha ao inicializar banco de dados: %v", err)
	}
	defer func() {
		log.Println("🔌 Fechando conexão com banco de dados...")
		if err := db.Close(); err != nil {
			log.Printf("❌ Erro ao fechar conexão: %v", err)
		}
	}()

	// Configura Fiber
	app := fiber.New(fiber.Config{
		AppName:      "Todo API v1.0",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		BodyLimit:    2 * 1024 * 1024, // 2MB
		ErrorHandler: globalErrorHandler,
	})

	// Middlewares globais
	setupMiddlewares(app)

	// Health check com estatísticas do banco
	app.Get("/health", createHealthCheckHandler(db))

	// Status do banco (endpoint para monitoramento)
	app.Get("/status", createStatusHandler(db))

	// Rotas da API
	api := app.Group("/api/v1")

	// Middleware para injetar database nas rotas
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})

	// Registra todas as rotas
	setupRoutes(api, db)

	// Graceful shutdown
	setupGracefulShutdown(app, db)

	// Inicia o servidor
	startServer(app, cfg.Port)
}

// setupMiddlewares configura todos os middlewares
func setupMiddlewares(app *fiber.App) {
	// Logger
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} - ${latency} | IP: ${ip}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	// Recover from panics
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // Em produção, configure domínios específicos
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
	}))
}

// globalErrorHandler trata erros globais da aplicação
func globalErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Erro interno do servidor"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	log.Printf("❌ Erro na API: %v | Path: %s | Method: %s", err, c.Path(), c.Method())

	return c.Status(code).JSON(fiber.Map{
		"success":   false,
		"error":     true,
		"message":   message,
		"code":      code,
		"timestamp": time.Now().Unix(),
		"path":      c.Path(),
	})
}

// createHealthCheckHandler cria handler para health check
func createHealthCheckHandler(db database.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Testa conexão com banco
		if err := db.Health(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":    "unhealthy",
				"timestamp": time.Now().Unix(),
				"service":   "todo-api",
				"version":   "1.0.0",
				"database":  "disconnected",
				"error":     err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status":      "healthy",
			"timestamp":   time.Now().Unix(),
			"service":     "todo-api",
			"version":     "1.0.0",
			"database":    "connected",
			"environment": os.Getenv("ENV"),
		})
	}
}

// createStatusHandler cria handler para status detalhado
func createStatusHandler(db database.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Pega estatísticas do banco
		stats, err := database.DatabaseStats(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Erro ao obter estatísticas do banco",
			})
		}

		return c.JSON(fiber.Map{
			"status":      "running",
			"timestamp":   time.Now().Unix(),
			"database":    stats,
			"environment": os.Getenv("ENV"),
			"uptime":      time.Now().Unix(),
		})
	}
}

// setupRoutes configura todas as rotas da aplicação
func setupRoutes(api fiber.Router, db database.Client) {
	// Rota de teste
	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":   "pong",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// Registrar rotas:
	// auth := api.Group("/auth")
	// users := api.Group("/users")
	// todos := api.Group("/todos")

	// E chamar os handlers:
	// handlers.SetupAuthRoutes(auth, db)
	// handlers.SetupUserRoutes(users, db)
	// handlers.SetupTodoRoutes(todos, db)
}

// setupGracefulShutdown configura shutdown gracioso
func setupGracefulShutdown(app *fiber.App, db database.Client) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("🔄 Iniciando graceful shutdown...")

		// Para de aceitar novas conexões
		if err := app.Shutdown(); err != nil {
			log.Printf("❌ Erro durante shutdown do servidor: %v", err)
		}
	}()
}

// startServer inicia o servidor
func startServer(app *fiber.App, port string) {
	log.Printf("🚀 Servidor rodando na porta %s", port)
	log.Printf("📊 Health check: http://localhost:%s/health", port)
	log.Printf("📈 Status: http://localhost:%s/status", port)
	log.Printf("📚 API Base: http://localhost:%s/api/v1", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
	}
}
