package gomplate

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	gotemplate "text/template"
	"time"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/properties"
	_ "github.com/flanksource/gomplate/v3/js"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/patrickmn/go-cache"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/registry"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/samber/oops"
	"google.golang.org/protobuf/types/known/structpb"
)

var funcMap gotemplate.FuncMap

var (
	// keep the cache period low as lots of anonymous functions can pile up the cache.
	goTemplateCache    = cache.New(time.Hour, time.Hour)
	celExpressionCache = cache.New(time.Hour, time.Hour)
)

func init() {
	funcMap = CreateFuncs(context.Background())
}

type Template struct {
	Template   string `yaml:"template,omitempty" json:"template,omitempty"` // Go template
	JSONPath   string `yaml:"jsonPath,omitempty" json:"jsonPath,omitempty"`
	Expression string `yaml:"expr,omitempty" json:"expr,omitempty"` // A cel-go expression
	Javascript string `yaml:"javascript,omitempty" json:"javascript,omitempty"`
	RightDelim string `yaml:"-" json:"-"`
	LeftDelim  string `yaml:"-" json:"-"`

	// DelimSets, when non-empty, runs the gotemplate over the input once per
	// delimiter pair, feeding the output of each pass into the next. Useful for
	// inputs that mix delimiter styles (e.g. {{ }} and $( )).
	// A header (# gotemplate: left-delim=… right-delim=…) overrides DelimSets to
	// a single pass. Falls back to LeftDelim/RightDelim if both are set, otherwise
	// to the default {{ }}.
	DelimSets []Delims `yaml:"-" json:"-"`

	// ValueFunctions, when true, exposes each key in the environment as a
	// zero-arg function in addition to dot-access. Enables {{ foo }} alongside
	// the standard {{ .foo }}.
	ValueFunctions bool `yaml:"-" json:"-"`

	// Pass in additional cel-env options like functions
	// that aren't simple enough to be included in Functions
	CelEnvs []cel.EnvOption `yaml:"-" json:"-"`

	// A map of functions that are accessible to cel expressions
	// and go templates.
	// NOTE: For cel expressions, the functions must be of type func() any.
	// If any other function type is used, an error will be returned.
	// Opt to CelEnvs for those cases.
	Functions map[string]any `yaml:"-" json:"-"`

	// CacheKey, when non-empty, is used as the program-cache key for this
	// template, bypassing the IsCacheable() heuristic. The caller asserts that
	// any two templates sharing this key may share the compiled cel.Program
	// (same expression semantics, same env shape, same function semantics).
	CacheKey string `yaml:"-" json:"-"`

	// CacheTime controls how long the compiled program/template is retained
	// in the cache. Zero means use the cache's default TTL; a positive value
	// is an explicit TTL; a negative value means no expiration.
	CacheTime time.Duration `yaml:"-" json:"-"`
}

func (t Template) String() string {
	if t.Template != "" {
		return "gotemplate: " + t.Template
	}
	if t.Expression != "" {
		return "cel: " + t.Expression
	}
	if t.Javascript != "" {
		return "js: " + t.Javascript
	}
	if t.JSONPath != "" {
		return "jsonpath: " + t.JSONPath
	}
	return ""
}

func (t Template) ShortString() string {
	if t.Template != "" {
		return "gotemplate: " + short(t.Template)
	}
	if t.Expression != "" {
		return "cel: " + short(t.Expression)
	}
	if t.Javascript != "" {
		return "js: " + short(t.Javascript)
	}
	if t.JSONPath != "" {
		return "jsonpath: " + short(t.JSONPath)
	}
	return ""
}

func short(v string) string {
	v = strings.TrimSpace(v)
	if len(v) == 0 {
		return ""
	}
	lines := strings.Split(v, "\n")
	if len(lines) == 1 {
		return lines[0]
	}
	return fmt.Sprintf("%s .. %d more lines", lines[0], len(lines)-1)
}

// autoCacheKey derives a cache key from the template fields and the env shape.
// Used when the caller has not supplied an explicit CacheKey.
func (t Template) autoCacheKey(env map[string]any) string {
	envVars := make([]string, 0, len(env)+1)
	for k := range env {
		envVars = append(envVars, k)
	}
	sort.Slice(envVars, func(i, j int) bool { return envVars[i] < envVars[j] })

	return strings.Join(envVars, "-") +
		t.RightDelim +
		t.LeftDelim +
		t.Expression +
		t.Javascript +
		t.JSONPath +
		t.Template
}

// cacheKey returns the key to use for the program/template cache. If the caller
// supplied CacheKey, it is used verbatim; otherwise the auto-derived key is used.
func (t Template) cacheKey(env map[string]any) string {
	if t.CacheKey != "" {
		return t.CacheKey
	}
	return t.autoCacheKey(env)
}

func (t Template) IsCacheable() bool {
	// An explicit CacheKey is the caller asserting the program is reusable
	// regardless of Functions/CelEnvs identity.
	if t.CacheKey != "" {
		return true
	}

	// Note: If custom functions are provided then we don't cache the template
	// because it's not possible to uniquely identify a function to be used as a cache key.
	// Pointers don't work well because different functions, that are behaviourly different,
	// but syntatically identical, will have the same pointer value.
	//
	// Reference: https://pkg.go.dev/reflect#Value.Pointer
	// 	> If v's Kind is Func, the returned pointer is an underlying code pointer,
	//  > but not necessarily enough to identify a single function uniquely.
	// 	> The only guarantee is that the result is zero if and only if v is a nil func Value.
	return len(t.CelEnvs) == 0 && len(t.Functions) == 0
}

func (t Template) IsEmpty() bool {
	return t.Template == "" && t.JSONPath == "" && t.Expression == "" && t.Javascript == ""
}

func RunExpression(_environment map[string]any, template Template) (any, error) {
	return RunExpressionContext(newContext(), _environment, template)
}

func RunExpressionContext(ctx commonsContext.Context, _environment map[string]any, template Template) (any, error) {
	data, err := Serialize(_environment)
	if err != nil {
		return "", err
	}

	// Look up the compiled-program cache BEFORE constructing the CEL env options.
	// GetCelEnv (notably kubernetes.Library()) is the dominant allocation on the CEL
	// path. On the overwhelmingly common cache hit it would be built and then
	// immediately discarded, since cel.NewEnv is only needed to compile a new
	// program. Build env options only when we actually need to compile.
	var prg cel.Program
	if template.IsCacheable() {
		cached, ok := celExpressionCache.Get(template.cacheKey(_environment))
		if ok {
			if cachedPrg, ok := cached.(*cel.Program); ok {
				prg = *cachedPrg
			}
		}
	}

	if prg == nil {
		base, err := baseCelEnv()
		if err != nil {
			return "", err
		}

		// Only the per-call options are layered on top of the cached base env: the
		// heavy, environment-independent libraries already live in base. This keeps
		// the dominant CEL setup cost (kubernetes.Library and declaration
		// validation) off the compile path.
		envOptions := make([]cel.EnvOption, 0, len(typeAdapters)+len(data)+len(template.Functions)+len(template.CelEnvs))
		envOptions = append(envOptions, typeAdapters...)
		for k := range data {
			envOptions = append(envOptions, cel.Variable(k, cel.AnyType))
		}
		for name, fn := range template.Functions {
			_name := name
			_fn := fn
			envOptions = append(envOptions, cel.Function(_name, cel.Overload(
				_name,
				nil,
				cel.AnyType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					ogFunc, ok := _fn.(func() any)
					if !ok {
						return types.WrapErr(fmt.Errorf("%s is expected to be of type func() any", _name))
					}

					out := ogFunc()
					return types.DefaultTypeAdapter.NativeToValue(out)
				}),
			)))
		}

		envOptions = append(envOptions, template.CelEnvs...)

		env, err := base.Extend(envOptions...)
		if err != nil {
			return "", err
		}

		ast, issues := env.Compile(strings.ReplaceAll(template.Expression, "\n", " "))
		if issues != nil && issues.Err() != nil {
			return "", oops.With("template", template.Expression).Errorf("issues: %s", issues.String())
		}

		prg, err = env.Program(ast)
		if err != nil {
			return "", err
		}

		if template.IsCacheable() {
			celExpressionCache.Set(template.cacheKey(_environment), &prg, template.CacheTime)
		}
	}

	out, _, err := prg.Eval(data)
	if err != nil {
		return nil, oops.With("template", template.Expression).Wrap(err)
	}
	if ctx.Logger != nil && out.Value() != template.Expression && properties.On(false, "gomplate.log") {
		ctx.Logger.V(4).Infof("templated %s => %v", template.ShortString(), out)
	}
	return out.Value(), nil

}

func newContext() commonsContext.Context {
	return commonsContext.NewContext(context.TODO(),
		commonsContext.WithLogger(logger.GetLogger("gomplate")))
}

func RunTemplateBool(environment map[string]any, template Template) (bool, error) {
	output, err := RunTemplateContext(newContext(), environment, template)
	if err != nil {
		return false, err
	}

	result, err := strconv.ParseBool(output)
	if err != nil {
		return false, fmt.Errorf("failed to parse template output (%s) as bool: %w", output, err)
	}

	return result, nil
}

func RunTemplate(environment map[string]any, template Template) (string, error) {
	return RunTemplateContext(newContext(), environment, template)
}

func RunTemplateContext(ctx commonsContext.Context, environment map[string]any, template Template) (string, error) {
	// javascript
	if template.Javascript != "" {
		vm := otto.New()
		for k, v := range environment {
			if err := vm.Set(k, v); err != nil {
				return "", fmt.Errorf("error setting %s: %w", k, err)
			}
		}

		out, err := vm.Run(template.Javascript)
		if err != nil {
			return "", fmt.Errorf("failed to run javascript: %w", err)
		}

		if s, err := out.ToString(); err != nil {
			return "", fmt.Errorf("failed to cast output to string: %w", err)
		} else {
			return s, nil
		}
	}

	// gotemplate
	if template.Template != "" {
		return runGoTemplate(ctx, template, environment)
	}

	// cel-go
	if template.Expression != "" {
		out, err := RunExpressionContext(ctx, environment, template)
		if err != nil {
			return "", err
		}
		if _, ok := out.(structpb.NullValue); ok || out == nil {
			return "", nil
		}
		return fmt.Sprintf("%v", out), nil
	}

	return "", nil
}

func runGoTemplate(ctx commonsContext.Context, template Template, environment map[string]any) (string, error) {
	// Parse the gotemplate header once up-front so we can detect whether the
	// header set delimiters (which overrides DelimSets to a single pass).
	origLeft, origRight := template.LeftDelim, template.RightDelim
	parsed, err := parseAndStripTemplateHeader(template)
	if err != nil {
		return "", err
	}
	template = parsed
	headerSetDelims := template.LeftDelim != origLeft || template.RightDelim != origRight

	if template.ValueFunctions {
		funcs := make(map[string]any, len(template.Functions)+len(environment))
		for k, v := range template.Functions {
			funcs[k] = v
		}
		for k, v := range environment {
			_v := v
			funcs[k] = func() any { return _v }
		}
		template.Functions = funcs
	}

	var delimSets []Delims
	switch {
	case headerSetDelims:
		delimSets = []Delims{{Left: template.LeftDelim, Right: template.RightDelim}}
	case len(template.DelimSets) > 0:
		delimSets = template.DelimSets
	case template.LeftDelim != "" && template.RightDelim != "":
		delimSets = []Delims{{Left: template.LeftDelim, Right: template.RightDelim}}
	default:
		delimSets = []Delims{{Left: "{{", Right: "}}"}}
	}

	val := template.Template
	for _, d := range delimSets {
		pass := template
		pass.Template = val
		pass.LeftDelim = d.Left
		pass.RightDelim = d.Right
		pass.DelimSets = nil
		out, err := goTemplate(ctx, pass, environment)
		if err != nil {
			return out, err
		}
		val = out
	}
	return val, nil
}

func goTemplate(ctx commonsContext.Context, template Template, environment map[string]any) (string, error) {
	var tpl *gotemplate.Template

	if template.IsCacheable() {
		cached, ok := goTemplateCache.Get(template.cacheKey(nil))
		if ok {
			if cachedTpl, ok := cached.(*gotemplate.Template); ok {
				if ctx.Logger != nil && properties.On(false, "gomplate.log") {
					ctx.Logger.V(7).Infof("%s using cached template", template.ShortString())
				}
				tpl = cachedTpl
			}
		}
	}

	if tpl == nil {
		template, err := parseAndStripTemplateHeader(template)
		if err != nil {
			return "", err
		}

		tpl = gotemplate.New("")
		if template.LeftDelim != "" {
			tpl = tpl.Delims(template.LeftDelim, template.RightDelim)
		}

		funcs := make(map[string]any)
		for k, v := range funcMap {
			funcs[k] = v
		}
		for k, v := range template.Functions {
			funcs[k] = v
		}

		tpl, err = tpl.Funcs(funcs).Parse(template.Template)
		if err != nil {
			return "", oops.With("template", template.Template).Wrap(err)
		}

		if template.IsCacheable() {
			goTemplateCache.Set(template.cacheKey(nil), tpl, template.CacheTime)
		}
	}

	data, err := Serialize(environment)
	if err != nil {
		return "", oops.
			// With("environment", environment)
			Wrapf(err, "error serializing env")
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", oops.
			With("template", template.Template).
			// With("environment", environment).
			Wrap(err)
	}

	out := strings.TrimSpace(buf.String())
	if ctx.Logger != nil && out != template.Template && properties.On(false, "gomplate.log") {
		ctx.Logger.V(4).Infof("templated %s ==> %s", template.ShortString(), out)
	}
	return out, nil
}

// LoadSharedLibrary loads a shared library for Otto
func LoadSharedLibrary(source string) error {
	source = strings.TrimSpace(source)
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("failed to read shared library %s: %s", source, err)
	}

	registry.Register(func() string { return string(data) })
	return nil
}

func parseAndStripTemplateHeader(template Template) (Template, error) {
	header, content := extractHeaderAndContent(template.Template)
	if header == "" {
		return template, nil
	}

	template.Template = content

	fields := strings.Fields(header)
	for _, field := range fields {
		split := strings.SplitN(field, "=", 2)
		if len(split) != 2 {
			return template, fmt.Errorf("invalid header: %s", field)
		}

		switch split[0] {
		case "right-delim":
			template.RightDelim = split[1]
		case "left-delim":
			template.LeftDelim = split[1]
		}
	}

	return template, nil
}

const templateHeaderPrefix = "# gotemplate: "

func extractHeaderAndContent(template string) (string, string) {
	scanner := bufio.NewScanner(strings.NewReader(template))

	// Loop through headers.
	// There could be multiple, we look for the gotemplate header.
	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" {
			// Special case for yaml where the header might not start from the first line.
			continue
		}

		// end of headers.
		isHeader := strings.HasPrefix(line, "#")
		if !isHeader {
			break
		}

		if strings.HasPrefix(line, templateHeaderPrefix) {
			header := strings.TrimPrefix(line, templateHeaderPrefix)
			return header, strings.Replace(template, fmt.Sprintf("%s\n", line), "", 1)
		}
	}

	return "", template
}
