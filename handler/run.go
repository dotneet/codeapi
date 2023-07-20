package handler

import (
	"fmt"
	"net/http"

	"github.com/dotneet/codeapi/runner"
	"github.com/labstack/echo/v4"
)

type JsonRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

func (handlers *Handlers) Run(c echo.Context) error {
	req := JsonRequest{}
	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	runner := runner.NewPythonRunner(handlers.ContainerImageName, handlers.Bucket)
	code := req.Code
	fmt.Printf("```%s\n%s```\n\n", req.Language, code)
	result, err := runner.Run(code)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"output": result.Output,
		"images": result.ImageUrls,
		"run_id": result.RunId,
	})
}
