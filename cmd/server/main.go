package main

import (
	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/joho/godotenv/autoload"

	"github.com/bongnv/jackett-stremio/internal/addon"
)

type config struct {
	JackettURL         string `env:"JACKETT_URL"`
	JackettAPIKey      string `env:"JACKETT_API_KEY"`
	RealDebridAPIToken string `env:"REAL_DEBRID_API_TOKEN"`
}

func main() {
	cfg := config{}
	_ = env.Parse(&cfg)

	app := fiber.New()
	app.Use(cors.New())
	app.Use(recover.New())
	app.Use(logger.New())

	add := addon.New(
		addon.WithID("stremio.addon.jackett"),
		addon.WithName("Jackett"),
		addon.WithJackett(cfg.JackettURL, cfg.JackettAPIKey),
		addon.WithRealDebrid(cfg.RealDebridAPIToken),
	)

	app.Get("/manifest.json", add.HandleGetManifest)
	app.Get("/stream/:type/:id.json", add.HandleGetStreams)
	app.Get("/download/:torrentID/:fileID", add.HandleDownload)
	app.Head("/download/:torrentID/:fileID", add.HandleDownload)

	log.Fatal(app.Listen(":7000"))
}
