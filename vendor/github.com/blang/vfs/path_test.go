package vfs

import (
	"reflect"

	"testing"
)

func TestSplitPath(t *testing.T) {
	const PathSeperator = "/"
	if p := SplitPath("/", PathSeperator); !reflect.DeepEqual(p, []string{""}) {
		t.Errorf("Invalid path: %q", p)
	}
	if p := SplitPath("./test", PathSeperator); !reflect.DeepEqual(p, []string{".", "test"}) {
		t.Errorf("Invalid path: %q", p)
	}
	if p := SplitPath(".", PathSeperator); !reflect.DeepEqual(p, []string{"."}) {
		t.Errorf("Invalid path: %q", p)
	}
	if p := SplitPath("test", PathSeperator); !reflect.DeepEqual(p, []string{".", "test"}) {
		t.Errorf("Invalid path: %q", p)
	}
	if p := SplitPath("/usr/src/linux/", PathSeperator); !reflect.DeepEqual(p, []string{"", "usr", "src", "linux"}) {
		t.Errorf("Invalid path: %q", p)
	}
	if p := SplitPath("usr/src/linux/", PathSeperator); !reflect.DeepEqual(p, []string{".", "usr", "src", "linux"}) {
		t.Errorf("Invalid path: %q", p)
	}
}
