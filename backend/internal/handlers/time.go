package handlers

import (
	"penguin-backend/internal/models"
	"penguin-backend/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

type TimeHandler struct{}

func NewTimeHandler() *TimeHandler {
	return &TimeHandler{}
}

// ParseTime godoc
// @Summary      Parse time string
// @Description  Parse various date/time string formats
// @Tags         time
// @Accept       json
// @Produce      json
// @Param        request body models.TimeParseRequest true "Time string to parse"
// @Success      200 {object} models.TimeParseResponse "Successful response"
// @Failure      400 {object} map[string]string "Bad request"
// @Router       /time/parse [post]
func (h *TimeHandler) ParseTime(c *fiber.Ctx) error {
	var req models.TimeParseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
	}

	parsedTime, err := utils.ParseTime(req.TimeString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Failed to parse time",
			"message": err.Error(),
		})
	}

	parsedTimestamp := models.NewTimestamp(parsedTime)
	response := models.ConvertDateTimeInstant(parsedTimestamp)
	response.Original = req.TimeString
	return c.JSON(response)
}

// GetSupportedFormats godoc
// @Summary      Get supported time formats
// @Description  Get list of all supported date/time formats
// @Tags         time
// @Accept       json
// @Produce      json
// @Success      200 {object} models.SupportedFormatsResponse "Successful response"
// @Router       /time/formats [get]
func (th *TimeHandler) GetSupportedFormats(c *fiber.Ctx) error {
	formats := []models.TimeFormat{
		{Name: "RFC3339Nano", Pattern: time.RFC3339Nano, Example: "2006-01-02T15:04:05.999999999Z07:00"},
		{Name: "RFC3339", Pattern: time.RFC3339, Example: "2006-01-02T15:04:05Z07:00"},
		{Name: "ISO8601 with ns", Pattern: "2006-01-02T15:04:05.999999999", Example: "2006-01-02T15:04:05.123456789"},
		{Name: "ISO8601", Pattern: "2006-01-02T15:04:05", Example: "2006-01-02T15:04:05"},
		{Name: "DateTime", Pattern: "2006-01-02 15:04:05", Example: "2006-01-02 15:04:05"},
		{Name: "Date", Pattern: "2006-01-02", Example: "2006-01-02"},
		{Name: "CompactDate", Pattern: "20060102", Example: "20060102"},
		{Name: "SlashDate", Pattern: "2006/01/02", Example: "2006/01/02"},
		{Name: "DotDate", Pattern: "2006.01.02", Example: "2006.01.02"},
		{Name: "Special Compact", Pattern: "YYYY-MMDD", Example: "1971-0618"},
	}

	return c.JSON(models.SupportedFormatsResponse{
		Formats: formats,
	})
}
