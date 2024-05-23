package tfiber

import (
	"github.com/gofiber/fiber/v2"
)

func init() {
	fiber.SetParserDecoder(fiber.ParserConfig{
		IgnoreUnknownKeys: true,
		SetAliasTag:       "json",
		ZeroEmpty:         true,
	})
}
