package gomplate

import (
	"testing"

	_ "github.com/flanksource/gomplate/v3/js"
	_ "github.com/robertkrimen/otto/underscore"
)

func Test_hashFunction(t *testing.T) {
	type args struct {
		env  map[string]any
		expr string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple",
			args: args{
				expr: "{{.name}}{{.age}}",
				env:  map[string]any{"age": 19, "name": "james"},
			},
			want: "age-name--{{.name}}{{.age}}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cacheKey(tt.args.env, tt.args.expr); got != tt.want {
				t.Errorf("hashFunction() = %v, want %v", got, tt.want)
			}
		})
	}
}
