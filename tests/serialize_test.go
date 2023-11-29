package tests

import (
	"testing"
	"time"

	"github.com/flanksource/gomplate/v3"
	"github.com/flanksource/gomplate/v3/data"
	_ "github.com/flanksource/gomplate/v3/js"
	"github.com/google/go-cmp/cmp"
	_ "github.com/robertkrimen/otto/underscore"
)

func Test_serialize(t *testing.T) {
	tests := []struct {
		name    string
		in      map[string]any
		want    map[string]any
		wantErr bool
	}{

		{
			name: "duration",
			in: map[string]any{
				"r": time.Second * 100,
				"a": time.Minute,
			},
			want: map[string]any{
				"r": (time.Second * 100),
				"a": time.Minute,
			},
		},
		{
			name: "time",
			in: map[string]any{
				"r": testDateTime,
			},
			want: map[string]any{
				"r": testDateTime,
			},
		},

		{
			name: "floats",
			in: map[string]any{
				"r": float64(100.50),
			},
			want: map[string]any{
				"r": 100.50,
			},
		},
		{
			name: "bytes",
			in: map[string]any{
				"r": []byte("hello world"),
			},
			want: map[string]any{
				"r": "hello world",
			},
		},
		{
			name: "dates",
			in: map[string]any{
				"r": newFile().Modified,
			},
			want: map[string]any{
				"r": newFile().Modified,
			},
		},
		{
			name: "nested_pointers",
			in: map[string]any{
				"r": newFolderCheck(1),
			},
			want: map[string]any{
				"r": map[string]any{
					"files": []any{
						map[string]any{
							"name":     "test",
							"size":     int64(10),
							"mode":     "drwxr-xr-x",
							"modified": testDate,
						},
					},
					"newest": map[string]any{
						"mode":     "drwxr-xr-x",
						"modified": testDate,
						"name":     "test",
						"size":     int64(10),
					},
				},
			},
		},
		{name: "nil", in: nil, want: nil, wantErr: false},
		{name: "empty", in: map[string]any{}, want: map[string]any{}, wantErr: false},
		{
			name:    "simple - no struct tags",
			in:      map[string]any{"r": NoStructTag{Name: "Kathmandu", UPPER: "u"}},
			want:    map[string]any{"r": map[string]any{"Name": "Kathmandu", "UPPER": "u"}},
			wantErr: false,
		},
		{name: "simple - struct tags", in: map[string]any{"r": Address{City: "Kathmandu"}}, want: map[string]any{"r": map[string]any{"city_name": "Kathmandu"}}, wantErr: false},
		{
			name:    "nested struct",
			in:      map[string]any{"r": Person{Name: "Aditya", Address: &Address{City: "Kathmandu"}}},
			want:    map[string]any{"r": map[string]any{"name": "Aditya", "Address": map[string]any{"city_name": "Kathmandu"}}},
			wantErr: false,
		},
		{
			name: "slice of struct",
			in: map[string]any{
				"r": []Address{
					{City: "Kathmandu"},
					{City: "Lalitpur"},
				},
			},
			want: map[string]any{
				"r": []any{
					map[string]any{"city_name": "Kathmandu"},
					map[string]any{"city_name": "Lalitpur"},
				},
			},
			wantErr: false,
		},
		{
			name: "nested slice of struct",
			in: map[string]any{
				"r": Person{
					Name: "Aditya",
					Addresses: []Address{
						{City: "Kathmandu"},
						{City: "Lalitpur"},
					},
				},
			},
			want: map[string]any{
				"r": map[string]any{
					"name": "Aditya",
					"addresses": []any{
						map[string]any{"city_name": "Kathmandu"},
						map[string]any{"city_name": "Lalitpur"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "pointers",
			in: map[string]any{
				"r": &Address{
					City: "Bhaktapur",
				},
			},
			want: map[string]any{
				"r": map[string]any{
					"city_name": "Bhaktapur",
				},
			},
		},
		{
			name: "canary checker ctx.Environment",
			in: map[string]any{
				"r": &Address{
					City: "Bhaktapur",
				},
				"check": map[string]any{
					"name": "custom-check",
					"meta": map[string]any{
						"key": "value",
					},
				},
			},
			want: map[string]any{
				"r": map[string]any{
					"city_name": "Bhaktapur",
				},
				"check": map[string]any{
					"name": "custom-check",
					"meta": map[string]any{
						"key": "value",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gomplate.Serialize(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("%v", diff)
				return
			}

			_got, _ := data.ToJSONPretty("  ", got)
			_want, _ := data.ToJSONPretty("  ", tt.want)
			if _got != _want {
				t.Errorf("serialize() = \n%s\nwant\n %v", _got, _want)
			}
		})
	}
}
