package main

import (
	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/joho/godotenv/autoload"

	"github.com/bongnv/prowlarr-stremio/internal/addon"
)

type config struct {
	ProwlarrURL        string `env:"PROWLARR_URL"`
	ProwlarrAPIKey     string `env:"PROWLARR_API_KEY"`
	RealDebridAPIToken string `env:"REAL_DEBRID_API_TOKEN"`
}

func main() {
	cfg := config{}
	_ = env.Parse(&cfg)

	app := fiber.New()
	app.Use(cors.New())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	app.Use(logger.New())

	add := addon.New(
		addon.WithID("stremio.addon.prowlarr"),
		addon.WithName("Prowlarr"),
		addon.WithProwlarr(cfg.ProwlarrURL, cfg.ProwlarrAPIKey),
		addon.WithRealDebrid(cfg.RealDebridAPIToken),
	)

	app.Get("/manifest.json", add.HandleGetManifest)
	app.Get("/stream/:type/:id.json", add.HandleGetStreams)
	app.Get("/download/:torrentID/:fileID", add.HandleDownload)
	app.Head("/download/:torrentID/:fileID", add.HandleDownload)

	log.Fatal(app.Listen(":7000"))
}
