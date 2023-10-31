package gomplate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type Test struct {
	Template   string `template:"true"`
	NoTemplate string
	Inner      Inner
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
		StructTemplater: StructTemplater{
			RequiredTag: "template",
			Values: map[string]interface{}{
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
		StructTemplater: StructTemplater{
			DelimSets: []Delims{
				{Left: "{{", Right: "}}"},
				{Left: "$(", Right: ")"},
			},
			Values: map[string]interface{}{
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
		StructTemplater: StructTemplater{
			RequiredTag: "template",
			DelimSets: []Delims{
				{Left: "{{", Right: "}}"},
				{Left: "$(", Right: ")"},
			},
			Values: map[string]interface{}{
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
}

func TestMain(t *testing.T) {
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
