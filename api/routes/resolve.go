package routes

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/youngjae-lim/url-shortener/database"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url")

	rc := database.CreateRedisClient(0)
	defer rc.Close()

  value, err :=	rc.Get(database.ctx, url)
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short not found in the database"})
	} else if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.map{"error": "cannot connect to DB"})
  }

  rIncrement := database.CreateRedisClient(1)
  defer rIncrement.Close()

  _ = rIncrement.Incr(database.ctx, "counter")

  return c.Redirect(value, 301)
}
