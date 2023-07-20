package runner

import (
	"github.com/dotneet/codeapi/storage"
	"testing"
)

func TestRunner(t *testing.T) {
	runner := NewPythonRunner(
		"python_runner",
		storage.ImageBucket{
			Endpoint: "localhost:9000",
		})
	code := "print('hello world')"
	result, _ := runner.Run(code)
	if result.Output != "hello world\n" {
		t.Errorf("Expected %s, got %s", "hello world", result.Output)
	}
}
