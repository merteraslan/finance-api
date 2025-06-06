package tefas

import (
	"net/http"
	"strings"
	"time"

	"github.com/ahmethakanbesel/finance-api/scraper"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
)

type Api struct {
	app     *pocketbase.PocketBase
	service Service
}

func NewApi(service Service, app *pocketbase.PocketBase) *Api {
	return &Api{
		app:     app,
		service: service,
	}
}

func (a *Api) SetupRoutes(g *echo.Group) {
	g.GET("/tefas/funds/:symbol", a.getFundData, apis.ActivityLogger(a.app))
}

func (a *Api) getFundData(c echo.Context) error {
	ctx := c.Request().Context()

	symbol := strings.ToUpper(c.PathParam("symbol"))
	if len(symbol) < 3 {
		return c.String(http.StatusBadRequest, "symbol must be at least 3 characters")
	}

	startDateStr := c.QueryParam("startDate")
	if startDateStr == "" {
		return c.String(http.StatusBadRequest, "start date must be provided")
	}
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return err
	}

	endDateStr := c.QueryParam("endDate")
	if endDateStr == "" {
		endDateStr = time.Now().Format("2006-01-02")
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return err
	}

	currency := strings.ToUpper(c.QueryParam("currency"))
	if currency != "TRY" && currency != "USD" {
		return c.String(http.StatusBadRequest, "currency must be either TRY or USD")
	}

	records, err := a.service.GetAndSaveFundData(ctx, symbol, currency, startDate, endDate)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Tell the client we're closing the connection after this response
	resp := c.Response()
	resp.Header().Set("Connection", "close")

	format := c.QueryParam("format")
	if format == "csv" {
		// Explicitly set CSV content type
		resp.Header().Set(echo.HeaderContentType, "text/csv")
		csvString, err := scraper.PriceRecordstoCsv(records)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, csvString)
	}

	// Default to JSON
	return c.JSON(http.StatusOK, &scraper.ApiResponse{
		Message: "ok",
		Data:    records,
	})
}
