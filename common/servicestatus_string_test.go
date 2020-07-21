package common

import "testing"

func TestServiceStatus_String(t *testing.T) {
	var ss ServiceStatus
	ss = -1
	t.Log(ss.String())
	ss = 2
	t.Log(ss.String())
}
