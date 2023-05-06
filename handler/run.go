package handler

import (
	"net/http"

	"github.com/dotneet/codeapi/runner"
	"github.com/labstack/echo/v4"
)

type JsonRequest struct {
	Code string `json:"code"`
}

func Run(c echo.Context) error {
	req := JsonRequest{}
	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	runner := runner.NewRunner()
	reader, err := runner.Run("python_runner", req.Code)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Stream(http.StatusOK, "text/plain", reader)
}
