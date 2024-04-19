package gomplate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type Test struct {
	Template   string `template:"true"`
	NoTemplate string
	Inner      Inner
	JSONMap    map[string]any    `template:"true"`
	Labels     map[string]string `template:"true"`
	LabelsRaw  map[string]string
}

type Inner struct {
	Template   string `template:"true"`
	NoTemplate string
}

type test struct {
	name string
	StructTemplater
	Input, Output *Test
	Vars          map[string]string
}

var tests = []test{
	{
		name: "template and no template",
		StructTemplater: StructTemplater{
			RequiredTag: "template",
			Values: map[string]any{
				"msg": "world",
			},
		},
		Input: &Test{
			Template:   "hello {{.msg}}",
			NoTemplate: "hello {{.msg}}",
		},
		Output: &Test{
			Template:   "hello world",
			NoTemplate: "hello {{.msg}}",
		},
	},
	{
		name: "just template",
		StructTemplater: StructTemplater{
			DelimSets: []Delims{
				{Left: "{{", Right: "}}"},
				{Left: "$(", Right: ")"},
			},
			Values: map[string]any{
				"msg": "world",
			},
			ValueFunctions: true,
		},
		Input: &Test{
			Template: "hello $(msg)",
		},
		Output: &Test{
			Template: "hello world",
		},
	},
	{
		name: "template & no template but with maps",
		StructTemplater: StructTemplater{
			RequiredTag: "template",
			DelimSets: []Delims{
				{Left: "{{", Right: "}}"},
				{Left: "$(", Right: ")"},
			},
			Values: map[string]any{
				"name":    "James Bond",
				"colorOf": "eye",
				"color":   "blue",
				"code":    "007",
				"city":    "London",
				"country": "UK",
			},
			ValueFunctions: true,
		},
		Input: &Test{
			Template: "Special Agent - $(name)!",
			Labels: map[string]string{
				"address":           "{{city}}, {{country}}",
				"{{colorOf}} color": "light $(color)",
				"code":              "{{code}}",
				"operation":         "noop",
			},
			LabelsRaw: map[string]string{
				"address":           "{{city}}, {{country}}",
				"{{colorOf}} color": "light $(color)",
			},
		},
		Output: &Test{
			Template: "Special Agent - James Bond!",
			Labels: map[string]string{
				"address":   "London, UK",
				"eye color": "light blue",
				"code":      "007",
				"operation": "noop",
			},
			LabelsRaw: map[string]string{
				"address":           "{{city}}, {{country}}",
				"{{colorOf}} color": "light $(color)",
			},
		},
	},
	{
		name: "deeply nested map",
		StructTemplater: StructTemplater{
			RequiredTag:    "template",
			ValueFunctions: true,
			Values: map[string]any{
				"msg": "world",
			},
		},
		Input: &Test{
			Template: "{{msg}}",
			JSONMap: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "{{msg}}",
					},
					"j": []map[string]any{
						{
							"l": "{{msg}}",
						},
					},
				},
				"e": "hello {{msg}}",
			},
		},
		Output: &Test{
			Template: "world",
			JSONMap: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "world",
					},
					"j": []any{
						map[string]any{
							"l": "world",
						},
					},
				},
				"e": "hello world",
			},
		},
	},
	{
		name: "pod manifest",
		StructTemplater: StructTemplater{
			RequiredTag:    "template",
			ValueFunctions: true,
			Values: map[string]any{
				"msg": "world",
			},
		},
		Input: &Test{
			JSONMap: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name":      "httpbin-{{msg}}",
					"namespace": "development",
					"labels": map[string]any{
						"app": "httpbin",
					},
				},
				"spec": map[string]any{
					"containers": []any{
						map[string]any{
							"name":  "httpbin",
							"image": "kennethreitz/httpbin:latest",
							"ports": []any{
								map[string]any{
									"containerPort": 80,
								},
							},
						},
					},
				},
			},
		},
		Output: &Test{
			JSONMap: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name":      "httpbin-world",
					"namespace": "development",
					"labels": map[string]any{
						"app": "httpbin",
					},
				},
				"spec": map[string]any{
					"containers": []any{
						map[string]any{
							"name":  "httpbin",
							"image": "kennethreitz/httpbin:latest",
							"ports": []any{
								map[string]any{
									"containerPort": 80,
								},
							},
						},
					},
				},
			},
		},
	},
}

func TestStructTemplater(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i := test.Input
			if err := test.StructTemplater.Walk(i); err != nil {
				t.Error(err)
			} else if diff := cmp.Diff(i, test.Output); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
