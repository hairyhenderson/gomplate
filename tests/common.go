package tests

type NoStructTag struct {
	Name  string
	UPPER string
}

type Address struct {
	City string `json:"city_name"`
}

type Person struct {
	Name      string         `json:"name"`
	Age       int            `json:"age,omitempty"`
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
