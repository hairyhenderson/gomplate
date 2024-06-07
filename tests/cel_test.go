package tests

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/flanksource/gomplate/v3"
	"github.com/flanksource/gomplate/v3/kubernetes"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func panIf(err error) {
	if err != nil {
		panic(err)
	}
}

func executeTemplate(t *testing.T, i int, input string, expectedOutput any, environment map[string]any, jsonCompare bool) {
	out, err := gomplate.RunExpression(environment, gomplate.Template{Expression: input})
	panIf(err)

	if jsonCompare {
		assert.JSONEq(t, expectedOutput.(string), out.(string), fmt.Sprintf("Test:%d failed", i+1))
	} else {
		assert.EqualValues(t, expectedOutput, out, fmt.Sprintf("Test:%d failed", i+1))
	}
}

type Test struct {
	env        map[string]interface{}
	expression string
	out        string
}

func runTests(t *testing.T, tests []Test) {
	for _, tc := range tests {
		t.Run(tc.expression, func(t *testing.T) {
			out, err := gomplate.RunTemplate(tc.env, gomplate.Template{
				Expression: tc.expression,
			})

			assert.ErrorIs(t, nil, err)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestFunctions(t *testing.T) {
	funcs := map[string]any{
		"fn": func() any {
			return map[string]any{
				"a": "b",
				"c": 1,
			}
		},
	}

	out, err := gomplate.RunTemplate(map[string]interface{}{
		"hello": "hi",
	}, gomplate.Template{
		Expression: "hello + ' ' + fn().a",
		Functions:  funcs,
	})

	assert.ErrorIs(t, nil, err)
	assert.Equal(t, "hi b", out)
}

// unstructure marshalls a struct to and from JSON to remove any type details
func unstructure(o any) interface{} {
	data, err := json.Marshal(o)
	if err != nil {
		return nil
	}
	var out interface{}
	_ = json.Unmarshal(data, &out)
	return out
}

func TestCelAws(t *testing.T) {
	runTests(t, []Test{
		{nil, "aws.arnToMap('arn:aws:sns:eu-west-1:123:MMS-Topic').account", "123"},
		{map[string]interface{}{
			"x": []map[string]string{
				{"Name": "hello", "Value": "world"},
				{"Name": "John", "Value": "Doe"},
			},
		},
			"aws.fromAWSMap(x).hello", "world"},
	})
}

func TestCelBase64(t *testing.T) {
	runTests(t, []Test{
		{nil, "base64.encode(b'hello')", "aGVsbG8="},
		{nil, "string(base64.decode('aGVsbG8='))", "hello"},
	})
}

func TestCelColl(t *testing.T) {
	runTests(t, []Test{
		{nil, `Dict(['a','b', 'c', 'd']).a`, "b"},
		{nil, `Dict(['a','b', 'c', 'd']).c`, "d"},
		{nil, `Has(['a','b', 'c'], 'a')`, "true"},
		{nil, `Has(['a','b', 'c'], 'e')`, "false"},
	})
}

func TestCelCrypto(t *testing.T) {
	runTests(t, []Test{
		{nil, `crypto.SHA1('hello')`, "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{nil, `crypto.SHA224('hello')`, "ea09ae9cc6768c50fcee903ed054556e5bfc8347907f12598aa24193"},
		{nil, `crypto.SHA256('hello')`, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{nil, `crypto.SHA384('hello')`, "59e1748777448c69de6b800d7a33bbfb9ff1b463e44354c3553bcdb9c666fa90125a3c79f90397bdf5f6a13de828684f"},
		{nil, `crypto.SHA512('hello')`, "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
	})
}

func TestCelData(t *testing.T) {
	person := Person{
		Name:    "Aditya",
		Address: &Address{City: "Kathmandu"},
	}
	runTests(t, []Test{
		{map[string]interface{}{"i": newFolderCheck(1)}, "i.files[0].modified", testDate},
		{map[string]interface{}{"i": person}, "YAML(toYAML(i)).name", "Aditya"},
		// csv
		{nil, `CSV(["Alice,30", "Bob,31"])[0][0]`, "Alice"},
		// {nil, `data.CSVByRow(["sn,name,age,gender", "1,Alice,30,f", "2,Bob,31,m"])`, "Alice"},
		// {nil, `toCSV("first,second,third\n1,2,3\n4,5,6")`, "Alice"},
		// TOML doesn't use JSON tags,
		{map[string]interface{}{"i": person}, "TOML(toTOML(i)).Address.city_name", "Kathmandu"},
		{structEnv, `results.Address.city_name == "Kathmandu" && results.name == "Aditya"`, "true"},
		// Support structs as environment var (by default they are not)
		{structEnv, `results.Address.city_name == "Kathmandu" && results.name == "Aditya"`, "true"},
		{map[string]any{"results": junitEnv}, `results.passed`, "1"},
	})
}

func TestCelFilePath(t *testing.T) {
	testData := []Test{
		{nil, `filepath.Base("/home/flanksource/projects/gencel")`, "gencel"},
		{nil, `filepath.Clean("/foo/bar/../baz")`, "/foo/baz"},
		{nil, `filepath.Dir("/home/flanksource/projects/gencel")`, "/home/flanksource/projects"},
		{nil, `filepath.Ext("/opt/image.jpg")`, ".jpg"},
		{nil, `filepath.IsAbs("/opt/image.jpg")`, "true"},
		{nil, `filepath.IsAbs("projects/image.jpg")`, "false"},
		{nil, `filepath.Join(["/home/flanksource", "projects", "gencel"])`, "/home/flanksource/projects/gencel"},
		{nil, `filepath.Match("*.txt", "foo.json")`, "false"},
		{nil, `filepath.Match("*.txt", "foo.txt")`, "true"},
		{nil, `filepath.Rel("/foo/bar", "/foo/bar/baz")`, "baz"},
		{nil, `filepath.Split("/foo/bar/baz")`, "[/foo/bar/ baz]"},
	}

	runTests(t, testData)
}

func TestCelJSON(t *testing.T) {
	person := Person{
		Name:    "Aditya",
		Address: &Address{City: "Kathmandu"},
	}
	runTests(t, []Test{
		{nil, `dyn([{'name': 'John', 'age': 30}]).toJSON()`, `[{"age":30,"name":"John"}]`},
		{nil, `[{'name': 'John'}].toJSON()`, `[{"name":"John"}]`},
		{nil, `dyn({'name': 'John'}).toJSON()`, `{"name":"John"}`},
		{nil, `{'name': 'John'}.toJSON()`, `{"name":"John"}`},
		{nil, `1.toJSON()`, `1`},
		{map[string]interface{}{"i": person}, "i.toJSON().JSON().name", "Aditya"},
		{map[string]interface{}{"i": person}, `'["1", "2"]'.JSONArray()[0]`, "1"},
		{map[string]interface{}{"i": map[string]string{"name": "aditya"}}, `i.toJSON()`, `{"name":"aditya"}`},

		{nil, `'{"name": "John"}'.JSON().name`, `John`},
		{nil, `'{"name": "Alice", "age": 30}'.JSON().name`, `Alice`},
		{nil, `'[1, 2, 3, 4, 5]'.JSONArray()[0]`, `1`},
		{map[string]interface{}{"i": person}, "jq('.Address.city_name', i)", "Kathmandu"},
		{map[string]interface{}{"i": person}, "i.toJSONPretty('\t')", "{\n\t\"Address\": {\n\t\t\"city_name\": \"Kathmandu\"\n\t},\n\t\"name\": \"Aditya\"\n}"},
		{nil, "[\"Alice\", 30].toJSONPretty('\t')", "[\n\t\"Alice\",\n\t30\n]"},
		{nil, "{'name': 'aditya'}.toJSONPretty('\t')", "{\n\t\"name\": \"aditya\"\n}"},
	})
}

func TestCelExtensions(t *testing.T) {
	runTests(t, []Test{
		{nil, "url('https://example.com/').getHost()", "example.com"},
		{nil, "[1,2, 3].min()", "1"},
		{nil, "[1,2, 3].max()", "3"},
		{nil, "[1,2, 3].sum()", "6"},
		// Extensions
		{nil, `base64.encode(b"flanksource")`, "Zmxhbmtzb3VyY2U="},        // encoding lib
		{nil, `string(base64.decode("Zmxhbmtzb3VyY2U="))`, "flanksource"}, // encoding lib
		{nil, `math.greatest(-42.0, -21.5, -100.0)`, "-21.5"},             // math lib
		{nil, `"hello, world".replace("world", "team")`, "hello, team"},   // strings lib
		{nil, `sets.contains([1, 2, 3, 4], [2, 3])`, "true"},              // sets lib
		{nil, `[1,2,3,4].slice(1, 3)`, "[2 3]"},                           // lists lib
	})
}

func TestCelEncode(t *testing.T) {
	tests := []Test{
		{map[string]interface{}{"hello": "hello world ?"}, "urlencode(hello)", `hello+world+%3F`},
		{map[string]interface{}{"hello": "hello+world+%3F"}, "urldecode(hello)", `hello world ?`},
	}

	runTests(t, tests)
}

func TestCelJQ(t *testing.T) {
	person := Person{Name: "Aditya", Address: &Address{City: "Kathmandu"}}
	persons := []interface{}{
		Person{Name: "John", Address: &Address{City: "Kathmandu"}, Age: 20},
		Person{Name: "Jane", Address: &Address{City: "Nepal"}, Age: 30},
		Person{Name: "Jane", Address: &Address{City: "Kathmandu"}, Age: 35},
		Person{Name: "Harry", Address: &Address{City: "Kathmandu"}, Age: 40},
	}

	runTests(t, []Test{
		{map[string]interface{}{"i": person}, "jq('.Address.city_name', i)", "Kathmandu"},
		{map[string]interface{}{"i": persons}, "jq('.[] | .name', i)", "[John Jane Jane Harry]"},
		{map[string]interface{}{"i": unstructure(persons)}, "jq('.[] | .name', i)", "[John Jane Jane Harry]"},
		{map[string]interface{}{"i": persons}, `jq('
			. |
		  group_by(.Address.city_name)  |
			map({"city": .[0].Address.city_name,
				 "sum": map(.age) | add,
				 "count": map(.) | length
			})', i)
			.map(v, "%s=%d".format([v.city, int(v.sum)]))
			.join(" ")`, "Kathmandu=95 Nepal=30"},
	})
}

func TestCelLists(t *testing.T) {
	runTests(t, []Test{
		{nil, "['a','b', 'c'].join(',')", "a,b,c"},

		{nil, "['a', ['b','c'], 'd'].flatten().join(',')", "a,b,c,d"},
		{nil, "['b', 'a', 'c'].sort().join(',')", "a,b,c"},
		{nil, "['b', 'a', 'c'].sort().join(',')", "a,b,c"},
		{nil, "['a', 'a', 'b', 'c'].uniq().join(',')", "a,b,c"},
		{nil, "{'a': 1, 'b': 2}.keys().join(',')", "a,b"},
		{nil, "{'a': 1, 'b': 2}.values().sum()", "3"},
	})
}

func TestCelMath(t *testing.T) {
	runTests(t, []Test{
		{nil, `math.Add([1,2,3,4,5])`, "15"},
		{nil, `math.Sub(5,4)`, "1"},
		{nil, `math.Mul([1, 2, 3, 4])`, "24"},
		{nil, `math.Div(4, 2)`, "2"},
		{nil, `math.Rem(4, 3)`, "1"},
		{nil, `math.Pow(4, 2)`, "16"},
		{nil, `math.Seq([1, 5])`, "[1 2 3 4 5]"},
		{nil, `math.Seq([1, 6, 2])`, "[1 3 5]"},
		{nil, `math.Abs(-1)`, "1"},
		{nil, `math.greatest([1,2,3,4,5])`, "5"},
		{nil, `math.least([1,2,3,4,5])`, "1"},
		{nil, `math.Ceil(5.4)`, "6"},
		{nil, `math.Floor(5.6)`, "5"},
		{nil, `math.Round(5.6)`, "6"},
	})
}

func TestCelMaps(t *testing.T) {
	m := map[string]any{
		"x": map[string]any{
			"a": "b",
			"c": 1,
			"d": true,
			"f": map[string]any{
				"a": "5",
			},
		},
	}
	runTests(t, []Test{
		{m, "x.a", "b"},
		{m, "x.c", "1"},
		{m, "x.d", "true"},
		{m, "x.?e", "<nil>"},
		{m, "x.?f.?a", "5"},
		{m, "x.?f.?b", "<nil>"},
		{nil, "{'a': 'c'}.merge({'b': 'd'}).keys().join(',')", "a,b"},
		{nil, "{'a': '1', 'b': '2', 'c': '3'}.pick(['a', 'c']).keys()", "[a c]"},
		{nil, "{'a': '1', 'b': '2', 'c': '3'}.omit(['b']).keys()", "[a c]"},
		{map[string]interface{}{"x": map[string]string{
			"a": "1",
			"b": "2",
			"c": "3",
		}}, "x.pick(['a', 'c']).keys()", "[a c]"},
	})
}

func TestCelRandom(t *testing.T) {
	runTests(t, []Test{
		{nil, `random.ASCII(1).size()`, "1"},
		{nil, `random.Alpha(25).size()`, "25"},
		{nil, `random.AlphaNum(32).size()`, "32"},
		{nil, `random.String(4, ['a', 'd']).size()`, "4"},
		{nil, `random.String(4, []).size()`, "4"},
		{nil, `random.Item(['a', 'b', 'c']).size()`, "1"},
		{nil, `random.Number(['12', '12'])`, "12"},
		// {nil, `random.Float([1, 2])`, "1.03"},
	})
}

func TestCelRegex(t *testing.T) {
	runTests(t, []Test{
		{nil, `' asdsa A123 asdsd'.find(r"A\d{3}")`, "A123"},
		{nil, `regexp.Find("[0-9]+","/path/i-213213/")`, "213213"},
		{nil, `regexp.Replace("[0-9]", ".","/path/i-213213/")`, "/path/i-....../"},
		{nil, `"abc 123".find('[0-9]+')`, "123"},
		{nil, `"/path/i-213213/aasd".matches('/path/(.*)/')`, "true"},
		{nil, `"ABC-123 213213 DEF-456 dsadjkl 4234".findAll(r'\w{3}-\d{3}')`, "[ABC-123 DEF-456]"},
	})
}

func TestCelStrings(t *testing.T) {
	tests := []Test{
		// Methods
		{nil, "'KubernetesPod'.abbrev(1, 5)", "Ku..."},
		{nil, "'KubernetesPod'.abbrev(6)", "Kub..."},
		{nil, "'Now is the time for all good men'.abbrev(5, 20)", "...s the time for..."},
		{nil, "'the quick brown fox'.camelCase()", "theQuickBrownFox"},
		{nil, "'the quick brown fox'.charAt(2)", "e"},
		{nil, "'the quick brown fox'.contains('brown')", "true"},
		{nil, "'the quick brown fox'.endsWith('fox')", "true"},
		{nil, "'the quick brown fox'.indent('\t\t\t')", "\t\t\tthe quick brown fox"},
		{nil, `"this is a string: %s\nand an integer: %d".format(["str", 42])`, "this is a string: str\nand an integer: 42"},
		{nil, "'hello world'.indent('==')", "==hello world"},
		{nil, "'hello world'.indent(4, '-')", "----hello world"},
		{nil, "'the quick brown fox'.indexOf('quick')", "4"},
		{nil, "['hello', 'mellow'].join()", "hellomellow"},
		{nil, "'hello mellow'.kebabCase()", "hello-mellow"},
		{nil, "'hello mellow'.kebabCase()", "hello-mellow"},
		{nil, "'hello hello hello'.lastIndexOf('hello')", "12"},
		{nil, "'HeyThERE'.lowerAscii()", "heythere"},
		{nil, "strings.quote('hello')", `"hello"`},
		{nil, "'hello'.quote()", `"hello"`},
		{nil, "'hello'.repeat(2)", `hellohello`},
		{nil, "'hello'.replaceAll('l', 'f')", `heffo`},
		{nil, "'hello'.reverse()", `olleh`},
		{nil, "'hello'.reverse()", `olleh`},
		{nil, `"Hello$World".runeCount()`, `11`},
		{nil, `"rm -rf /home/*".shellQuote()`, `'rm -rf /home/*'`},
		{nil, `"hello".size()`, `5`},
		{nil, `"hello there".slug()`, `hello-there`},
		{map[string]interface{}{"s": "hello world"}, "s.snakeCase()", "hello_world"},
		{map[string]interface{}{"s": "hello world"}, "s.shellQuote()", "'hello world'"},
		{nil, `"hello".sort()`, `ehllo`},
		{map[string]interface{}{"s": "hello world"}, "s.split(' ')", "[hello world]"},
		{map[string]interface{}{"s": "hello world"}, "s.squote()", "'hello world'"},
		{nil, `"hello".startsWith("he")`, "true"},
		{map[string]interface{}{"s": "hello world"}, "s.title()", "Hello World"},
		{map[string]interface{}{"s": "    hello world\t\t\n"}, "s.trim()", "hello world"},
		{map[string]interface{}{"s": "hello world"}, "s.trimSuffix(' world')", "hello"},
		{map[string]interface{}{"s": "hello world"}, "s.slug()", "hello-world"},
		{map[string]interface{}{"s": "testing this line from here"}, "s.wordWrap(2)", "testing\nthis\nline\nfrom\nhere"},
		{map[string]interface{}{"s": "testing this line from here"}, "s.wordWrap(10)", "testing\nthis line\nfrom here"},
		{map[string]interface{}{"s": "Hello$World"}, "s.wordWrap(5)", "Hello$World"},
		{nil, `"Hello Beautiful World".wordWrap(16, '===')`, "Hello Beautiful===World"},
		{nil, `"Hello Beautiful World".wordWrap(25, '')`, "Hello Beautiful World"}, // no need to wrap

		// Functions
		{nil, "HumanSize(123456)", "120.6K"},
		{nil, "Semver('1.2.3').major", "1"},
		{map[string]interface{}{"v": "1.2.3-beta.1+c0ff33"}, "Semver(v).prerelease", "beta.1"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.3"}, "SemverCompare(new, old)", "true"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.4"}, "SemverCompare(new, old)", "false"},

		// Basic
		{map[string]interface{}{"hello": "world"}, "hello", "world"},
		{map[string]interface{}{"code": 200}, "string(code) + ' is not equal to 500'", "200 is not equal to 500"},
	}

	runTests(t, tests)
}

func TestCelDates(t *testing.T) {
	timestamp, err := time.Parse(time.RFC3339Nano, "2020-01-01T14:30:33.456Z")
	assert.NoError(t, err)
	tests := []Test{
		{map[string]interface{}{"t": timestamp}, "t.getSeconds()", "33"},
		{map[string]interface{}{"t": timestamp}, "string(t)", "2020-01-01T14:30:33.456Z"},

		// Durations
		{map[string]interface{}{"age": 75 * time.Second}, "age", "1m15s"},
		{nil, `HumanDuration(duration("1008h"))`, "6w0d0h"},
		{nil, `Duration("7d").getHours()`, "168"},
		{nil, `duration("1h") > duration("2h")`, "false"},
		{map[string]interface{}{"t": "2020-01-01T00:00:00Z"}, `Age(t) > duration('1h')`, "true"},
		// {map[string]interface{}{"t": "2020-01-01T00:00:00Z"}, `age(t) > duration('1h')`, "true"},
		{map[string]interface{}{"t": "2020-01-01T00:00:00Z"}, `Age(t) > Duration('3d')`, "true"},
		{nil, `duration('24h') > Duration('3d')`, "false"},

		// {map[string]interface{}{"code": 200, "sslAge": time.Hour}, `code in [200,201,301] && sslAge > duration('59m')`, "true"},

		{map[string]interface{}{"code": 200, "sslAge": time.Hour}, `sslAge`, "1h0m0s"},
		{map[string]interface{}{"code": 200, "sslAge": time.Hour}, `sslAge < duration('2h')`, "true"},
		{map[string]interface{}{"code": 200, "sslAge": time.Hour}, `code in [200,201,301] && sslAge > duration('59m')`, "true"},
	}

	runTests(t, tests)
}

func TestCelVariadic(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		{Input: `math.Add([1,2,3,4,5])`, Output: int64(15)},
		{Input: `math.Mul([1,2,3,4,5])`, Output: int64(120)},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil, false)
	}
}

func TestCelSliceReturn(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		{Input: `"open-source".split("-")`, Output: []string{"open", "source"}},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil, false)
	}
}

func TestCelK8sResources(t *testing.T) {
	runTests(t, []Test{
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructuredMap(kubernetes.TestHealthySvc)}, "k8s.isHealthy(healthySvc)", "true"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructuredMap(kubernetes.TestHealthySvc)}, "k8s.isReady(healthySvc)", "true"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructuredMap(kubernetes.TestLuaStatus)}, "k8s.getStatus(healthySvc)", ": found less than two generators, Merge requires two or more"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructuredMap(kubernetes.TestHealthySvc)}, "k8s.getHealth(healthySvc).health", "healthy"},
	})
}

func TestCelK8s(t *testing.T) {
	testData := []struct {
		Input       string
		Output      any
		jsonCompare bool
	}{
		{Input: `k8s.getHealth(healthy_obj).health`, Output: "healthy"},
		{Input: `k8s.isHealthy(healthy_obj)`, Output: true},
		{Input: `k8s.isHealthy(unhealthy_obj)`, Output: false},
		{Input: `k8s.getHealth(healthy_obj).status`, Output: "Running"},
		{Input: `k8s.getHealth(unhealthy_obj).message`, Output: "Back-off 40s restarting failed container=main pod=my-pod_argocd(63674389-f613-11e8-a057-fe5f49266390)"},
		{Input: `k8s.getHealth(unhealthy_obj).ok`, Output: false},
		{Input: `k8s.getHealth(healthy_obj).message`, Output: ""},
		{Input: `k8s.is_healthy(healthy_obj)`, Output: true},
		{Input: `dyn(obj_list).all(i, k8s.isHealthy(i))`, Output: false},
		{Input: `dyn(unstructured_list).all(i, k8s.isHealthy(i))`, Output: false},
		{Input: `k8s.isHealthy(unhealthy_obj)`, Output: false},
		{Input: `k8s.neat(service_raw)`, Output: kubernetes.TestServiceNeat, jsonCompare: true},
		{Input: `k8s.neat(pod_raw)`, Output: kubernetes.TestPodNeat, jsonCompare: true},
		{Input: `k8s.neat(pod_raw_obj.Object)`, Output: kubernetes.TestPodNeat, jsonCompare: true},
		{Input: `k8s.neat(pv_raw, 'yaml')`, Output: kubernetes.TestPVYAMLRaw},
	}

	var podRaw unstructured.Unstructured
	if err := json.Unmarshal([]byte(kubernetes.TestPodRaw), &podRaw.Object); err != nil {
		t.Fatal(err)
	}

	environment := map[string]any{
		"healthy_obj":       kubernetes.TestHealthySvc,
		"unhealthy_obj":     kubernetes.TestUnhealthy,
		"obj_list":          []string{kubernetes.TestHealthySvc, kubernetes.TestUnhealthy},
		"unstructured_list": kubernetes.TestUnstructuredList,
		"service_raw":       kubernetes.TestServiceRaw,
		"pod_raw":           kubernetes.TestPodRaw,
		"pod_raw_obj":       podRaw,
		"pv_raw":            kubernetes.TestPVJsonRaw,
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, environment, td.jsonCompare)
	}
}

func TestCelK8sCPUResourceUnits(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		{Input: `k8s.cpuAsMillicores("10m")`, Output: int64(10)},
		{Input: `k8s.cpuAsMillicores("100m")`, Output: int64(100)},
		{Input: `k8s.cpuAsMillicores("1000m")`, Output: int64(1000)},
		{Input: `k8s.cpuAsMillicores("15n")`, Output: int64(0)},
		{Input: `k8s.cpuAsMillicores("150n")`, Output: int64(0)},
		{Input: `k8s.cpuAsMillicores("15000000n")`, Output: int64(15)},
		{Input: `k8s.cpuAsMillicores("150000000n")`, Output: int64(150)},
		{Input: `k8s.cpuAsMillicores("0.5")`, Output: int64(500)},
		{Input: `k8s.cpuAsMillicores("1")`, Output: int64(1000)},
		{Input: `k8s.cpuAsMillicores("1.5")`, Output: int64(1500)},
		{Input: `k8s.cpuAsMillicores("1.234")`, Output: int64(1234)},
		{Input: `k8s.cpuAsMillicores("5")`, Output: int64(5000)},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil, false)
	}
}

func TestCelK8sMemoryResourceUnits(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		{Input: `k8s.memoryAsBytes("10Ki")`, Output: int64(10240)},
		{Input: `k8s.memoryAsBytes("100Ki")`, Output: int64(102400)},
		{Input: `k8s.memoryAsBytes("1000Ki")`, Output: int64(1024000)},
		{Input: `k8s.memoryAsBytes("50Mi")`, Output: int64(52428800)},
		{Input: `k8s.memoryAsBytes("500Mi")`, Output: int64(524288000)},
		{Input: `k8s.memoryAsBytes("512Mi")`, Output: int64(536870912)},
		{Input: `k8s.memoryAsBytes("1Gi")`, Output: int64(1073741824)},
		{Input: `k8s.memoryAsBytes("1.234Gi")`, Output: int64(1324997410)},
		{Input: `k8s.memoryAsBytes("5Gi")`, Output: int64(5368709120)},
		{Input: `k8s.memoryAsBytes("10ki")`, Output: int64(10240)},
		{Input: `k8s.memoryAsBytes("100ki")`, Output: int64(102400)},
		{Input: `k8s.memoryAsBytes("1000ki")`, Output: int64(1024000)},
		{Input: `k8s.memoryAsBytes("50mi")`, Output: int64(52428800)},
		{Input: `k8s.memoryAsBytes("500mi")`, Output: int64(524288000)},
		{Input: `k8s.memoryAsBytes("512mi")`, Output: int64(536870912)},
		{Input: `k8s.memoryAsBytes("1gi")`, Output: int64(1073741824)},
		{Input: `k8s.memoryAsBytes("1.234gi")`, Output: int64(1324997410)},
		{Input: `k8s.memoryAsBytes("5gi")`, Output: int64(5368709120)},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil, false)
	}
}

func TestCelYAML(t *testing.T) {
	person := Person{
		Name:    "Aditya",
		Address: &Address{City: "Kathmandu"},
	}
	runTests(t, []Test{
		{nil, "YAML('name: John\n').name", "John"},
		{nil, "YAMLArray('- name\n')[0]", "name"},
		{nil, `toYAML(['name', 'John'])`, "- name\n- John\n"},
		{map[string]any{"person": person}, `toYAML(person)`, "Address:\n  city_name: Kathmandu\nname: Aditya\n"},
		{nil, `toYAML({'name': 'John'})`, "name: John\n"},
	})
}

func TestCelTOML(t *testing.T) {
	person := Person{
		Name: "Aditya",
	}

	runTests(t, []Test{
		{nil, "TOML('name = \"John\"').name", "John"},
		{map[string]any{"person": person}, `toTOML(person)`, "name = \"Aditya\"\n"},
	})
}

func TestCelUUID(t *testing.T) {
	runTests(t, []Test{
		{nil, "uuid.Nil()", "00000000-0000-0000-0000-000000000000"},
		{nil, "uuid.V1() != uuid.Nil()", "true"},
		{nil, "uuid.V4() != uuid.Nil()", "true"},
		{nil, "uuid.IsValid('2a42e576-c308-4db9-8525-0513af307586')", "true"},
		{nil, "uuid.Parse('2a42e576-c308-4db9-8525-0513af307586')", "2a42e576-c308-4db9-8525-0513af307586"},
	})
}
