package template

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type VariableReplacer struct {
	APIBase string
}

type responseBodyInterceptor struct {
	http.ResponseWriter
	body *bytes.Buffer
	ctx  echo.Context
	code int
}

func (r *responseBodyInterceptor) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *responseBodyInterceptor) WriteHeader(code int) {
	r.code = code
}

// replace the variables in the response body.
// this middleware should be used only for the request to the files in ".well-known" directory.
func (replacer *VariableReplacer) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// when the request path does not contain ".well-known" do nothing.
		if !bytes.Contains([]byte(c.Request().URL.Path), []byte("/.well-known/")) {
			return next(c)
		}

		// Wrap the response writer with responseBodyInterceptor
		res := c.Response()
		interceptor := &responseBodyInterceptor{
			ResponseWriter: res.Writer,
			body:           new(bytes.Buffer),
			ctx:            c,
			code:           200,
		}
		res.Writer = interceptor

		// Call the next handler in the chain
		err := next(c)
		if err != nil {
			return err
		}

		// Modify the response body
		originalBody, _ := ioutil.ReadAll(interceptor.body)
		modifiedBody := bytes.ReplaceAll(originalBody, []byte("${api_base}"), []byte(replacer.APIBase))

		// Write the modified response body
		modifiledLen := strconv.Itoa(len(modifiedBody))
		interceptor.ResponseWriter.Header().Set(echo.HeaderContentLength, modifiledLen)
		interceptor.ResponseWriter.WriteHeader(interceptor.code)
		_, err = interceptor.ResponseWriter.Write(modifiedBody)
		return err
	}
}
