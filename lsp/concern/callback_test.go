package concern

import (
	"testing"
)

func TestDefaultCallback(t *testing.T) {
	var d DefaultCallback
	d.NotifyBeforeCallback(nil)
	d.NotifyAfterCallback(nil, nil)
}
