package runner

import (
	"github.com/dotneet/codeapi/storage"
	"testing"
)

func TestRunner(t *testing.T) {
	runner := NewPythonRunner(storage.ImageBucket{
		Endpoint: "localhost:9000",
	})
	code := "print('hello world')"
	result, _ := runner.Run("python_runner", code)
	if result.Output != "hello world\n" {
		t.Errorf("Expected %s, got %s", "hello world", result.Output)
	}
}
