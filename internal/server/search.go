package server

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/search"
)

func (s *Server) handleSearch(c echo.Context) error {
	q := c.QueryParam("q")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	resp, err := search.Run(database, q, limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}
