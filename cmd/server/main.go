package main

import (
	"os"
	"regexp"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/joho/godotenv/autoload"

	"github.com/bongnv/prowlarr-stremio/internal/addon"
	"github.com/bongnv/prowlarr-stremio/internal/static"
)

type config struct {
	ProwlarrURL    string `env:"PROWLARR_URL"`
	ProwlarrAPIKey string `env:"PROWLARR_API_KEY"`
	Production     bool   `env:"PRODUCTION"`
}

var (
	maskedPathPattern = regexp.MustCompile(`^/([\w%]+)/(?:configure|stream|download|manifest)`)
	version           = "0.1.0-dev"
)

func main() {
	cfg := config{}
	_ = env.Parse(&cfg)

	app := fiber.New()
	app.Use(cors.New())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(logger.New(logger.Config{
		CustomTags: map[string]logger.LogFunc{
			"maskedPath": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				urlPath := c.Path()
				loc := maskedPathPattern.FindStringSubmatchIndex(urlPath)
				if len(loc) > 1 {
					return output.WriteString(urlPath[:loc[0]+1] + "***" + urlPath[loc[1]:])
				} else {
					return output.WriteString(urlPath)
				}
			},
		},
		Format:        "${time} | ${status} | ${latency} | ${ip} | ${method} | ${maskedPath} | ${error}\n",
		TimeFormat:    "15:04:05",
		TimeZone:      "Local",
		TimeInterval:  500 * time.Millisecond,
		Output:        os.Stdout,
		DisableColors: false,
	}))

	add := addon.New(
		addon.WithID("stremio.addon.prowlarr"),
		addon.WithName("Prowlarr"),
		addon.WithProwlarr(cfg.ProwlarrURL, cfg.ProwlarrAPIKey),
		addon.WithDevelopment(!cfg.Production),
		addon.WithVersion(version),
	)

	app.Get("/manifest.json", add.HandleGetManifest)
	app.Get("/:userData/manifest.json", add.HandleGetManifest)
	app.Get("/:userData/stream/:type/:id.json", add.HandleGetStreams)
	app.Get("/:userData/download/:infoHash/:fileID", add.HandleDownload)
	app.Head("/:userData/download/:infoHash/:fileID", add.HandleDownload)
	app.Get("/configure", static.HandleConfigure)
	app.Get("/:userData/configure", static.HandleConfigure)

	log.Fatal(app.Listen(":7000"))
}
