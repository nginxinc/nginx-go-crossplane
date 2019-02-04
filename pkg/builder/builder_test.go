package builder

import (
	"testing"
)

func TestBuild(t *testing.T) {
	c, err := Build("test", 4, false, false)
	if err != nil {
		t.Errorf("test failed due to error being returned from Build %s", err.Error())
	}
	if c != "built" {
		t.Errorf("expected %s but got %s", "built", c)
	}
}
