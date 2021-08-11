package concern_manager

import "testing"

func TestDefaultCallback(t *testing.T) {
	var d defaultCallback
	d.NotifyBeforeCallback(nil)
	d.NotifyAfterCallback(nil, nil)
}
