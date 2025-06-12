package main

import (
	"log"
	"penguin-backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	_ "penguin-backend/docs"

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title Penguin Folder Management API
// @version 1.0.0
// @description API for managing and browsing folders
// @host localhost:8080
// @BasePath /api
func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	folderHandler := handlers.NewFolderHandler()
	timeHandler := handlers.NewTimeHandler()

	api := app.Group("/api")
	api.Get("/folders", folderHandler.GetFolders)
	api.Get("/kouji-projects", folderHandler.GetKoujiProjects)
	api.Post("/kouji-projects/save", folderHandler.SaveKoujiProjectsToYAML)
	api.Put("/kouji-projects/:project_id/dates", folderHandler.UpdateKoujiProjectDates)
	api.Post("/kouji-projects/cleanup", folderHandler.CleanupInvalidTimeData)
	api.Post("/time/parse", timeHandler.ParseTime)
	api.Get("/time/formats", timeHandler.GetSupportedFormats)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Penguin Backend API",
			"version": "1.0.0",
			"docs":    "/swagger/index.html",
		})
	})

	log.Println("Server starting on :8080")
	log.Println("API documentation available at http://localhost:8080/swagger/index.html")
	log.Fatal(app.Listen(":8080"))
}
