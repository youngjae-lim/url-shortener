package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/youngjae-lim/url-shortener/database"
	"github.com/youngjae-lim/url-shortener/helpers"
)

// request is a type struct for HTTP request
type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

// response is a type struct for HTTP response
type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

// ShortenURL shortens a user-enter
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
	val, err := rc2.Get(database.Ctx, c.IP()).Result()

	if err == redis.Nil {
		// set the quota using .env and TTL to 30 minutes if the request is for the first time
		_ = rc2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot connect to DB"})
	}

	// val, _ = rc2.Get(database.Ctx, c.IP()).Result()
	valInt, _ := strconv.Atoi(val)

	if valInt <= 0 {
		// get the remaining time to live of the key
		limit, _ := rc2.TTL(database.Ctx, c.IP()).Result()

		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":            "Rate limit exceeded",
			"rate_limit_reset": limit / time.Nanosecond / time.Minute,
		})
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

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	// new redis client
	r := database.CreateRedisClient(0)
	defer r.Close()

	// check if id is already in use
	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "URL custom short is already in use"})
	}

	// set expiry in the request body
	if body.Expiry == 0 { // if there is no expiry set
		body.Expiry = 24 // 24 hours
	}

	// set 24 hours expiry in the database
	err = r.Set(database.Ctx, id, body.URL, body.Expiry*60*60*time.Second).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to connect to server",
		})
	}

	// construct response
	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10, // 10 api quota
		XRateLimitReset: 30, // every 30 minutes
	}
	// decrement api quota by 1 for the requesting ip
	rc2.Decr(database.Ctx, c.IP())

	// get the decreased rate remaining
	val, _ = rc2.Get(database.Ctx, c.IP()).Result()

	// update XRateRemaining with the decreased rate limit remaining
	resp.XRateRemaining, _ = strconv.Atoi(val)

	// get the TTL in milisecond
	// 1s = 10^9 nanosecond
	// 1min = 60s = 60 * 10^9 nanosecond
	// for example, 30 mins = 30 * 60 * 10^9 nanosecond
	ttl, _ := rc2.TTL(database.Ctx, c.IP()).Result()

	// update XRateLimitReset with the current TTL
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	// set CustomShort
	// example: bit.ly/2kkxjLk
	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
