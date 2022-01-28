package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/youngjae-lim/url-shortener/database"
	"github.com/youngjae-lim/url-shortener/helpers"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)

	// bind the request body to a struct
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// implement rate limiting
	rc2 := database.CreateRedisClient(1)
	defer rc2.Close()

  // get current api quota for the requesting ip
	val, err := rc2.Get(database.ctx, c.IP()).Result()

	if err == redis.Nil {
    // set the quota to 30 minutes if the request is for the first time
		_ = rc2.Set(database.ctx, c.IP(), os.Getenv("API_QUOTA"), 30*time.Minute).Err()
	} else if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.map{"error": "cannot connect to DB"})
  } else {
		// val, _ = rc2.Get(database.ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)

    if valInt <= 0 {
      // get the remaining time to live of the key
      limit, _ := rc2.TTL(database.ctx, c.IP()).Result()
      return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.map{
        "error": "Rate limit exceeded",
        "rate_limit_reset": (limit * time.Minute) / time.Nanosecond
      })
    }
	}

	// check if the input is an actual URL
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	// check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "you can't hack the system"})
	}

	// enforce https, SSL
	body.URL = helpers.EnforceHTTP(body.URL)

  // decrement api quota by 1 for the requesting ip
  rc2.Decr(database.ctx, c.IP())

	return nil
}
