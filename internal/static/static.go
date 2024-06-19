package static

import (
	_ "embed"

	"github.com/gofiber/fiber/v2"
)

//go:embed configure.html
var configure []byte

func HandleConfigure(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return c.Send(configure)
}
