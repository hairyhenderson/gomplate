package tests

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/flanksource/gomplate/v3"
	"github.com/flanksource/gomplate/v3/kubernetes"
	"github.com/stretchr/testify/assert"
)

func panIf(err error) {
	if err != nil {
		panic(err)
	}
}

func executeTemplate(t *testing.T, i int, input string, expectedOutput any, environment map[string]any) {
	out, err := gomplate.RunExpression(environment, gomplate.Template{Expression: input})
	panIf(err)
	assert.EqualValues(t, expectedOutput, out, fmt.Sprintf("Test:%d failed", i+1))
}

type Test struct {
	env        map[string]interface{}
	expression string
	out        string
}

func TestCel(t *testing.T) {

	tests := []Test{}

	runTests(t, tests)
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

func TestRegex(t *testing.T) {
	runTests(t, []Test{
		{nil, `' asdsa A123 asdsd'.find(r"A\d{3}")`, "A123"},
		{nil, `regexp.Find("[0-9]+","/path/i-213213/")`, "213213"},
		{nil, `regexp.Replace("[0-9]", ".","/path/i-213213/")`, "/path/i-....../"},
		{nil, `"abc 123".find('[0-9]+')`, "123"},
		{nil, `"/path/i-213213/aasd".matches('/path/(.*)/')`, "true"},
		{nil, `"ABC-123 213213 DEF-456 dsadjkl 4234".findAll(r'\w{3}-\d{3}')`, "[ABC-123 DEF-456]"},
	})
}

func TestMath(t *testing.T) {
	runTests(t, []Test{
		{nil, `math.Add([1,2,3,4,5])`, "15"},
	})
}

func TestMaps(t *testing.T) {
	runTests(t, []Test{

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

func TestLists(t *testing.T) {
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

func TestAws(t *testing.T) {
	runTests(t, []Test{
		{nil, "arnToMap('arn:aws:sns:eu-west-1:123:MMS-Topic').account", "123"},
		{map[string]interface{}{"x": []map[string]string{
			{"Name": "hello", "Value": "world"},
			{"Name": "John", "Value": "Doe"},
		}}, "fromAWSMap(x).hello", "world"},
	})
}

func TestJQ(t *testing.T) {
	person := Person{Name: "Aditya", Address: &Address{City: "Kathmandu"}}
	persons := []interface{}{
		Person{Name: "John", Address: &Address{City: "Kathmandu"}, Age: 20},
		Person{Name: "Jane", Address: &Address{City: "Nepal"}, Age: 30},
		Person{Name: "Jane", Address: &Address{City: "Kathmandu"}, Age: 35},

		Person{Name: "Harry", Address: &Address{City: "Kathmandu"}, Age: 40},
	}

	runTests(t, []Test{
		{map[string]interface{}{"i": person}, "jq('.Address.city_name', i)", "Kathmandu"},
		{map[string]interface{}{"i": persons}, "jq('.[] | .name', i)", "[John Jane]"},
		{map[string]interface{}{"i": persons}, "jq('.', i).toJSON()", "[John Jane]"},
		{map[string]interface{}{"i": unstructure(persons)}, "jq('.[] | .name', i)", "[John Jane]"},
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

func TestData(t *testing.T) {
	person := Person{Name: "Aditya", Address: &Address{City: "Kathmandu"}}
	runTests(t, []Test{
		// {map[string]interface{}{"i": person}, "jq('.address.city_name', i)", "Aditya"},
		{map[string]interface{}{"i": person}, "toJSONPretty(i)", "{\n  \"Address\": {\n    \"city_name\": \"Kathmandu\"\n  },\n  \"name\": \"Aditya\"\n}"},
		{map[string]interface{}{"i": person}, "JSON(toJSON(i)).name", "Aditya"},
		{map[string]interface{}{"i": person}, "YAML(toYAML(i)).name", "Aditya"},
		{map[string]interface{}{"i": person}, "TOML(toTOML(i)).name", "Aditya"},

		{structEnv, `results.Address.city_name == "Kathmandu" && results.name == "Aditya"`, "true"},
		// Support structs as environment var (by default they are not)
		{structEnv, `results.Address.city_name == "Kathmandu" && results.name == "Aditya"`, "true"},
		{map[string]any{"results": junitEnv}, `results.passed`, "1"},
	})
}

func TestExtensions(t *testing.T) {
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
func TestStrings(t *testing.T) {
	tests := []Test{
		{nil, "random.String(10, ['a','b','d'])", ""},
		{map[string]interface{}{"hello": "world"}, "hello", "world"},
		{map[string]interface{}{"hello": "hello world ?"}, "urlencode(hello)", `hello+world+%3F`},
		{map[string]interface{}{"hello": "hello+world+%3F"}, "urldecode(hello)", `hello world ?`},
		// {map[string]interface{}{"size": "123456"}, "HumanSize(size)", "120.6K"},
		{map[string]interface{}{"size": 123456}, "HumanSize(size)", "120.6K"},
		{map[string]interface{}{"size": 123456}, "humanSize(size)", "120.6K"},
		{map[string]interface{}{"v": "1.2.3-beta.1+c0ff33"}, "Semver(v).prerelease", "beta.1"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.3"}, "SemverCompare(new, old)", "true"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.4"}, "SemverCompare(new, old)", "false"},
		{map[string]interface{}{"code": 200}, "string(code) + ' is not equal to 500'", "200 is not equal to 500"},

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
		{nil, "uuid.IsValid(uuid.V1())", "true"},
		{nil, "uuid.IsValid(uuid.V4())", "true"},
	}

	runTests(t, tests)
}

func TestDates(t *testing.T) {
	tests := []Test{
		// Durations
		{map[string]interface{}{"age": 75 * time.Second}, "age", "1m15s"},
		{nil, `HumanDuration(duration("1008h"))`, "6w0d0h"},
		{nil, `humanDuration(duration("1008h"))`, "6w0d0h"},
		{nil, `Duration("7d").getHours()`, "168"},
		{nil, `duration("1h") > duration("2h")`, "false"},
		{map[string]interface{}{"t": "2020-01-01T00:00:00Z"}, `Age(t) > duration('1h')`, "true"},
		// {map[string]interface{}{"t": "2020-01-01T00:00:00Z"}, `age(t) > duration('1h')`, "true"},
		{map[string]interface{}{"t": "2020-01-01T00:00:00Z"}, `Age(t) > Duration('3d')`, "true"},
		{nil, `duration('24h') > Duration('3d')`, "false"},
		{map[string]interface{}{"code": 200, "sslAge": time.Hour}, `code in [200,201,301] && sslAge > duration('59m')`, "true"},
	}

	runTests(t, tests)
}

func TestCelNamespace(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		// {Input: `regexp.Replace("flank", "rank", "flanksource")`, Output: "ranksource"},
		// {Input: `regexp.Replace("nothing", "rank", "flanksource")`, Output: "flanksource"},
		// {Input: `regexp.Replace("", "", "flanksource")`, Output: "flanksource"},
		{Input: `filepath.Join(["/home/flanksource", "projects", "gencel"])`, Output: "/home/flanksource/projects/gencel"},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil)
	}
}

func TestCelMultipleReturns(t *testing.T) {
	testData := []struct {
		Input   string
		Outputs []any
	}{
		// {Input: `base64.Encode("flanksource")`, Outputs: []any{"Zmxhbmtzb3VyY2U=", nil}},
		// {Input: `base64.Decode("Zmxhbmtzb3VyY2U=")`, Outputs: []any{"flanksource", nil}},
		{Input: `JSONArray("[\"name\",\"flanksource\"]")`, Outputs: []any{"name", "flanksource"}},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Outputs, nil)
	}
}

func TestCelVariadic(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		{Input: `math.Add([1,2,3,4,5])`, Output: int64(15)},
		{Input: `math.Mul([1,2,3,4,5])`, Output: int64(120)},
		{Input: `Slice([1,2,3,4,5])`, Output: []any{int64(1), int64(2), int64(3), int64(4), int64(5)}},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil)
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
		executeTemplate(t, i, td.Input, td.Output, nil)
	}
}

func TestCelK8sResources(t *testing.T) {

	runTests(t, []Test{
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructuredMap(kubernetes.TestHealthy)}, "IsHealthy(healthySvc)", "true"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructuredMap(kubernetes.TestLuaStatus)}, "GetStatus(healthySvc)", "Degraded: found less than two generators, Merge requires two or more"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructuredMap(kubernetes.TestHealthy)}, "GetHealth(healthySvc).status", "Healthy"},
	})
}

func TestCelK8s(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{

		{Input: `k8s.is_healthy(healthy_obj)`, Output: true},
		{Input: `k8s.is_healthy(unhealthy_obj)`, Output: false},
		{Input: `k8s.health(healthy_obj).status`, Output: "Healthy"},
		{Input: `k8s.health(unhealthy_obj).message`, Output: "Back-off 40s restarting failed container=main pod=my-pod_argocd(63674389-f613-11e8-a057-fe5f49266390)"},
		{Input: `k8s.health(unhealthy_obj).ok`, Output: false},
		{Input: `k8s.health(healthy_obj).message`, Output: ""},
	}

	environment := map[string]any{
		"healthy_obj":   kubernetes.TestHealthy,
		"unhealthy_obj": kubernetes.TestUnhealthy,
	}
	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, environment)
	}
}

func TestCelJSON(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		{Input: `dyn([{'name': 'John', 'age': 30}]).toJSON()`, Output: `[{"age":30,"name":"John"}]`},
		{Input: `[{'name': 'John'}].toJSON()`, Output: `[{"name":"John"}]`},
		{Input: `dyn({'name': 'John'}).toJSON()`, Output: `{"name":"John"}`},
		{Input: `{'name': 'John'}.toJSON()`, Output: `{"name":"John"}`},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil)
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
		{Input: `k8s.cpuAsMillicores("0.5")`, Output: int64(500)},
		{Input: `k8s.cpuAsMillicores("1")`, Output: int64(1000)},
		{Input: `k8s.cpuAsMillicores("1.5")`, Output: int64(1500)},
		{Input: `k8s.cpuAsMillicores("1.234")`, Output: int64(1234)},
		{Input: `k8s.cpuAsMillicores("5")`, Output: int64(5000)},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil)
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
		executeTemplate(t, i, td.Input, td.Output, nil)
	}
}
