package gomplate

import (
	"reflect"
	"testing"
	"time"

	_ "github.com/flanksource/gomplate/v3/js"
	"github.com/flanksource/gomplate/v3/k8s"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/stretchr/testify/assert"
)

type NoStructTag struct {
	Name  string
	UPPER string
}

type Address struct {
	City string `json:"city_name"`
}

type Person struct {
	Name      string         `json:"name"`
	Address   *Address       `json:",omitempty"`
	MetaData  map[string]any `json:",omitempty"`
	Codes     []string       `json:",omitempty"`
	Addresses []Address      `json:"addresses,omitempty"`
}

// A shared test data for all template test
var structEnv = map[string]any{
	"results": Person{
		Name: "Aditya",
		Address: &Address{
			City: "Kathmandu",
		},
	},
}

type Totals struct {
	Passed   int     `json:"passed"`
	Failed   int     `json:"failed"`
	Skipped  int     `json:"skipped,omitempty"`
	Error    int     `json:"error,omitempty"`
	Duration float64 `json:"duration"`
}

type JunitTest struct {
	Name string `json:"name" yaml:"name"`
}

type JunitTestSuite struct {
	Name   string `json:"name"`
	Totals `json:",inline"`
	Tests  []JunitTest `json:"tests"`
}

type JunitTestSuites struct {
	Suites []JunitTestSuite `json:"suites,omitempty"`
	Totals `json:",inline"`
}

var junitEnv = JunitTestSuites{
	Totals: Totals{
		Passed: 1,
	},
	Suites: []JunitTestSuite{{Name: "hi", Totals: Totals{Failed: 2}}},
}

type SQLDetails struct {
	Rows  []map[string]interface{} `json:"rows,omitempty"`
	Count int                      `json:"count,omitempty"`
}

func TestJavascript(t *testing.T) {
	tests := []struct {
		env map[string]interface{}
		js  string
		out string
	}{
		{map[string]interface{}{"x": 5, "y": 3}, "x + y", "8"},
		{map[string]interface{}{"str": "Hello, World!"}, "str", "Hello, World!"},
		{map[string]interface{}{"numbers": []int{1, 2, 3, 4, 5}}, "_.reduce(numbers, function(memo, num){ return memo + num; }, 0)", "15"},
		{map[string]interface{}{"numbers": []int{4, 2, 55}}, `_.max(numbers)`, "55"},
		{map[string]interface{}{"arr": []int{1, 2, 1, 4, 1, 2}}, "_.uniq(arr)", "1,2,4"},
		{map[string]interface{}{"numbers": []int{4, 2, 55}}, `_.max(numbers)`, "55"},
		{map[string]interface{}{"arr": []int{1, 2, 1, 4, 1, 2}}, "_.uniq(arr)", "1,2,4"},
		{map[string]interface{}{"x": "1Ki"}, "fromSI(x)", "1024"},
		{map[string]interface{}{"x": "2m"}, "fromMillicores(x)", "2"},
		{map[string]interface{}{"name": "mission.compute.internal"}, "k8s.getNodeName(name)", "mission"},
		{map[string]interface{}{"msg": map[string]any{"healthStatus": map[string]string{"status": "HEALTHY"}}}, "k8s.conditions.isReady(msg)", "true"},
		{structEnv, `results.name + " " + results.Address.city_name`, "Aditya Kathmandu"},
	}

	for _, tc := range tests {
		t.Run(tc.js, func(t *testing.T) {
			out, err := RunTemplate(tc.env, Template{
				Javascript: tc.js,
			})
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestGomplate(t *testing.T) {
	tests := []struct {
		env      map[string]interface{}
		template string
		out      string
	}{
		{map[string]interface{}{"hello": "world"}, "{{ .hello }}", "world"},
		{map[string]interface{}{"hello": "hello world ?"}, "{{ .hello | urlencode }}", `hello+world+%3F`},
		{map[string]interface{}{"hello": "hello+world+%3F"}, "{{ .hello | urldecode }}", `hello world ?`},
		{map[string]interface{}{"age": 75 * time.Second}, "{{ .age | humanDuration  }}", "1m15s"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructured(k8s.TestHealthy)}, "{{ (.healthySvc | isHealthy) }}", "true"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructured(k8s.TestLuaStatus)}, "{{ (.healthySvc | getStatus) }}", "Degraded: found less than two generators, Merge requires two or more"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructured(k8s.TestHealthy)}, "{{ (.healthySvc | getHealth).Status  }}", "Healthy"},
		{map[string]interface{}{"size": 123456}, "{{ .size | humanSize }}", "120.6K"},
		{map[string]interface{}{"v": "1.2.3-beta.1+c0ff33"}, "{{  (.v | semver).Prerelease  }}", "beta.1"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.3"}, "{{  .old | semverCompare .new }}", "true"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.4"}, "{{  .old | semverCompare .new }}", "false"},
		{structEnv, `{{.results.name}} {{.results.Address.city_name}}`, "Aditya Kathmandu"},
		{
			map[string]any{"results": junitEnv},
			`{{.results.passed}}{{ range $r := .results.suites}}{{$r.name}} ✅ {{$r.passed}} ❌ {{$r.failed}} in 🕑 {{$r.duration}}{{end}}`,
			"1hi ✅ 0 ❌ 2 in 🕑 0",
		},
		{
			map[string]any{
				"results": SQLDetails{
					Rows: []map[string]any{{"name": "apm-hub"}, {"name": "config-db"}},
				},
			},
			`{{range $r := .results.rows }}{{range $x, $y := $r }}{{ $y }}{{end}}{{end}}`, "apm-hubconfig-db"},
	}

	for _, tc := range tests {
		t.Run(tc.template, func(t *testing.T) {
			out, err := RunTemplate(tc.env, Template{
				Template: tc.template,
			})
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestCel(t *testing.T) {
	person := Person{Name: "Aditya", Address: &Address{City: "Kathmandu"}}
	tests := []struct {
		env        map[string]interface{}
		expression string
		out        string
	}{

		{nil, "uuid.IsValid(uuid.V1())", "true"},
		{nil, "uuid.IsValid(uuid.V4())", "true"},
		{nil, "[1,2, 3].min()", "1"},
		{nil, "[1,2, 3].max()", "3"},
		{nil, "[1,2, 3].sum()", "6"},
		{nil, "{'a': 1, 'b': 2}.keys().join(',')", "a,b"},
		{nil, "{'a': 1, 'b': 2}.values().sum()", "3"},

		// {map[string]interface{}{"i": person}, "jq('.address.city_name', i)", "Aditya"},
		{map[string]interface{}{"i": person}, "toJSONPretty(i)", "{\n  \"Address\": {\n    \"city_name\": \"Kathmandu\"\n  },\n  \"name\": \"Aditya\"\n}"},
		{map[string]interface{}{"i": person}, "JSON(toJSON(i)).name", "Aditya"},
		{map[string]interface{}{"i": person}, "YAML(toYAML(i)).name", "Aditya"},
		{map[string]interface{}{"i": person}, "TOML(toTOML(i)).name", "Aditya"},

		// {map[string]interface{}{"s": []string{"a", "b", "b", "a"}}, "Uniq(s)", "[a b]"},
		{map[string]interface{}{"s": "hello world"}, "s.title()", "Hello World"},
		{map[string]interface{}{"s": "hello world"}, "s.camelCase()", "helloWorld"},
		{map[string]interface{}{"s": "hello world"}, "s.kebabCase()", "hello-world"},
		{map[string]interface{}{"s": "hello world"}, "s.snakeCase()", "hello_world"},
		{map[string]interface{}{"s": "hel\"lo world"}, "s.quote()", "\"hel\\\"lo world\""},
		{map[string]interface{}{"s": "hello world"}, "s.squote()", "'hello world'"},
		{map[string]interface{}{"s": "hello world"}, "s.shellQuote()", "'hello world'"},
		{map[string]interface{}{"s": "hello world"}, "s.slug()", "hello-world"},
		// {map[string]interface{}{"s": "hello world"}, "s.runeCount()", "Hello World"},

		{nil, `math.Add([1,2,3,4,5])`, "15"},
		{map[string]interface{}{"hello": "world"}, "hello", "world"},
		{map[string]interface{}{"hello": "hello world ?"}, "urlencode(hello)", `hello+world+%3F`},
		{map[string]interface{}{"hello": "hello+world+%3F"}, "urldecode(hello)", `hello world ?`},
		{map[string]interface{}{"age": 75 * time.Second}, "age", "1m15s"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructuredMap(k8s.TestHealthy)}, "IsHealthy(healthySvc)", "true"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructuredMap(k8s.TestLuaStatus)}, "GetStatus(healthySvc)", "Degraded: found less than two generators, Merge requires two or more"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructuredMap(k8s.TestHealthy)}, "GetHealth(healthySvc).status", "Healthy"},
		// {map[string]interface{}{"size": "123456"}, "HumanSize(size)", "120.6K"},
		{map[string]interface{}{"size": 123456}, "HumanSize(size)", "120.6K"},
		{map[string]interface{}{"size": 123456}, "humanSize(size)", "120.6K"},
		{map[string]interface{}{"v": "1.2.3-beta.1+c0ff33"}, "Semver(v).prerelease", "beta.1"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.3"}, "SemverCompare(new, old)", "true"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.4"}, "SemverCompare(new, old)", "false"},
		{map[string]interface{}{"code": 200}, "string(code) + ' is not equal to 500'", "200 is not equal to 500"},

		// Durations
		{nil, `HumanDuration(duration("1008h"))`, "6w0d0h"},
		{nil, `humanDuration(duration("1008h"))`, "6w0d0h"},

		{nil, `Duration("7d").getHours()`, "168"},
		{nil, `duration("1h") > duration("2h")`, "false"},
		{map[string]interface{}{"t": "2020-01-01T00:00:00Z"}, `Age(t) > duration('1h')`, "true"},
		// {map[string]interface{}{"t": "2020-01-01T00:00:00Z"}, `age(t) > duration('1h')`, "true"},
		{map[string]interface{}{"t": "2020-01-01T00:00:00Z"}, `Age(t) > Duration('3d')`, "true"},
		{nil, `duration('24h') > Duration('3d')`, "false"},
		{map[string]interface{}{"code": 200, "sslAge": time.Hour}, `code in [200,201,301] && sslAge > duration('59m')`, "true"},

		// Extensions
		{nil, `base64.encode(b"flanksource")`, "Zmxhbmtzb3VyY2U="},        // encoding lib
		{nil, `string(base64.decode("Zmxhbmtzb3VyY2U="))`, "flanksource"}, // encoding lib
		{nil, `math.greatest(-42.0, -21.5, -100.0)`, "-21.5"},             // math lib
		{nil, `"hello, world".replace("world", "team")`, "hello, team"},   // strings lib
		{nil, `sets.contains([1, 2, 3, 4], [2, 3])`, "true"},              // sets lib
		{nil, `[1,2,3,4].slice(1, 3)`, "[2 3]"},                           // lists lib

		{structEnv, `results.Address.city_name == "Kathmandu" && results.name == "Aditya"`, "true"},
		// Support structs as environment var (by default they are not)
		{structEnv, `results.Address.city_name == "Kathmandu" && results.name == "Aditya"`, "true"},
		{map[string]any{"results": junitEnv}, `results.passed`, "1"},
	}

	for _, tc := range tests {
		t.Run(tc.expression, func(t *testing.T) {
			out, err := RunTemplate(tc.env, Template{
				Expression: tc.expression,
			})
			assert.ErrorIs(t, nil, err)
			assert.Equal(t, tc.out, out)
		})
	}
}

func Test_serialize(t *testing.T) {
	tests := []struct {
		name    string
		in      map[string]any
		want    map[string]any
		wantErr bool
	}{
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
				"r": []map[string]any{
					{"city_name": "Kathmandu"},
					{"city_name": "Lalitpur"},
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
					"addresses": []map[string]any{
						{"city_name": "Kathmandu"},
						{"city_name": "Lalitpur"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := serialize(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serialize() = %v, want %v", got, tt.want)
			}
		})
	}
}
