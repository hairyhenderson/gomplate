package coll

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/client-go/util/jsonpath"
)

type missingKeyCtxKey struct{}

// ContextWithMissingKey returns a context with the missing-key policy set.
func ContextWithMissingKey(ctx context.Context, missingKey string) context.Context {
	return context.WithValue(ctx, missingKeyCtxKey{}, missingKey)
}

// MissingKeyFromContext returns the missing-key policy from the context.
// Defaults to "error" if not set.
func MissingKeyFromContext(ctx context.Context) string {
	if ctx == nil {
		return "error"
	}
	if v, ok := ctx.Value(missingKeyCtxKey{}).(string); ok && v != "" {
		return v
	}
	return "error"
}

// JSONPath -
func JSONPath(p string, in any) (any, error) {
	return JSONPathWithContext(context.Background(), p, in)
}

// JSONPathWithContext -
func JSONPathWithContext(ctx context.Context, p string, in any) (any, error) {
	missingKey := MissingKeyFromContext(ctx)
	allowMissingKeys := missingKey == "zero" || missingKey == "default" || missingKey == "invalid"

	jp, err := parsePath(p, allowMissingKeys)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse JSONPath %s: %w", p, err)
	}
	results, err := jp.FindResults(in)
	if err != nil {
		return handleFindErr(ctx, p, in, err, allowMissingKeys)
	}

	if len(results) == 1 && len(results[0]) == 1 {
		return extractResult(results[0][0])
	}

	a, err := collectResults(results)
	if err != nil {
		return nil, err
	}
	if allowMissingKeys && len(a) == 0 {
		return nil, nil
	}
	return a, nil
}

func handleFindErr(ctx context.Context, p string, in any, err error, allowMissingKeys bool) (any, error) {
	if !allowMissingKeys || !strings.Contains(err.Error(), "array index out of bounds") {
		return nil, fmt.Errorf("executing JSONPath failed: %w", err)
	}
	matches, ok, recErr := evalMultiBranchBounds(ctx, p, in)
	if recErr != nil {
		return nil, fmt.Errorf("executing JSONPath failed: %w", recErr)
	}
	if ok {
		return matches, nil
	}
	return nil, nil
}

func evalMultiBranchBounds(ctx context.Context, p string, in any) (any, bool, error) {
	missingKey := MissingKeyFromContext(ctx)
	allowMissingKeys := missingKey == "zero" || missingKey == "default" || missingKey == "invalid"

	raw, ok, err := evalMultiBranchRaw(ctx, p, in, allowMissingKeys)
	if !ok || err != nil {
		return nil, ok, err
	}
	if len(raw) == 0 {
		return nil, true, nil
	}

	matches := make([]any, 0, len(raw))
	for _, v := range raw {
		res, err := extractResult(v)
		if err != nil {
			return nil, true, err
		}
		if res != nil {
			matches = append(matches, res)
		}
	}

	if len(matches) == 0 {
		return nil, true, nil
	}
	if len(matches) == 1 {
		return matches[0], true, nil
	}
	return matches, true, nil
}

func evalMultiBranchRaw(ctx context.Context, p string, in any, allowMissingKeys bool) ([]reflect.Value, bool, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, true, ctx.Err()
	}

	prefixPath, suffixPath, ok := splitBranchNodes(p)
	if !ok {
		return nil, false, nil
	}

	branchItems, err := evalPrefixBranches(in, prefixPath, allowMissingKeys)
	if err != nil {
		return nil, true, err
	}
	if len(branchItems) == 0 {
		return nil, true, nil
	}

	var aggregated []reflect.Value
	for _, item := range branchItems {
		raw, err := evalBranch(ctx, suffixPath, item, allowMissingKeys)
		if err != nil {
			return nil, true, err
		}
		aggregated = append(aggregated, raw...)
	}

	return aggregated, true, nil
}

func splitBranchNodes(p string) (prefixPath, suffixPath string, ok bool) {
	parser, err := jsonpath.Parse("<jsonpath>", "{"+p+"}")
	if err != nil || len(parser.Root.Nodes) == 0 {
		return "", "", false
	}
	listNode, okType := parser.Root.Nodes[0].(*jsonpath.ListNode)
	if !okType || len(listNode.Nodes) == 0 {
		return "", "", false
	}

	branchIdx := -1
	for i, n := range listNode.Nodes {
		if isBranchingNode(n) {
			branchIdx = i
			break
		}
	}
	if branchIdx == -1 {
		return "", "", false
	}

	prefixNodes := listNode.Nodes[:branchIdx+1]
	suffixNodes := listNode.Nodes[branchIdx+1:]
	if len(suffixNodes) == 0 {
		if _, isUnion := listNode.Nodes[branchIdx].(*jsonpath.UnionNode); !isUnion {
			return "", "", false
		}
	}

	prefixPath, err = formatNodes(prefixNodes)
	if err != nil {
		return "", "", false
	}
	if len(suffixNodes) > 0 {
		suffixPath, err = formatNodes(suffixNodes)
		if err != nil {
			return "", "", false
		}
	}
	return prefixPath, suffixPath, true
}

func evalPrefixBranches(in any, prefixPath string, allowMissingKeys bool) ([]any, error) {
	parser, err := jsonpath.Parse("<jsonpath>", "{"+prefixPath+"}")
	if err != nil || len(parser.Root.Nodes) == 0 {
		return nil, nil
	}
	listNode, okType := parser.Root.Nodes[0].(*jsonpath.ListNode)
	if !okType || len(listNode.Nodes) == 0 {
		return nil, nil
	}

	lastIdx := len(listNode.Nodes) - 1
	if unionNode, isUnion := listNode.Nodes[lastIdx].(*jsonpath.UnionNode); isUnion {
		baseNodes := listNode.Nodes[:lastIdx]
		basePath, err := formatNodes(baseNodes)
		if err != nil {
			return nil, err
		}
		if basePath == "$" || basePath == "." {
			basePath = ""
		}

		var branchItems []any
		for _, child := range unionNode.Nodes {
			elemStr, err := formatUnionElement(child.Nodes)
			if err != nil {
				return nil, err
			}
			opPath := basePath + "[" + elemStr + "]"
			opJp, err := parsePath(opPath, allowMissingKeys)
			if err != nil {
				return nil, err
			}
			opResults, err := opJp.FindResults(in)
			if err != nil {
				if allowMissingKeys && strings.Contains(err.Error(), "array index out of bounds") {
					continue
				}
				return nil, err
			}
			for _, row := range opResults {
				for _, v := range row {
					if !v.IsValid() {
						continue
					}
					if !v.CanInterface() {
						return nil, fmt.Errorf("JSONPath couldn't access field")
					}
					branchItems = append(branchItems, v.Interface())
				}
			}
		}
		return branchItems, nil
	}

	prefixJp, err := parsePath(prefixPath, allowMissingKeys)
	if err != nil {
		return nil, nil
	}
	prefixResults, err := prefixJp.FindResults(in)
	if err != nil {
		if !allowMissingKeys || !strings.Contains(err.Error(), "array index out of bounds") {
			return nil, err
		}
		return nil, nil
	}

	var branchItems []any
	for _, row := range prefixResults {
		for _, v := range row {
			if !v.IsValid() {
				continue
			}
			if !v.CanInterface() {
				return nil, fmt.Errorf("JSONPath couldn't access field")
			}
			branchItems = append(branchItems, v.Interface())
		}
	}
	return branchItems, nil
}

func evalBranch(ctx context.Context, suffixPath string, item any, allowMissingKeys bool) ([]reflect.Value, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if suffixPath == "" {
		return []reflect.Value{reflect.ValueOf(item)}, nil
	}

	subJp, err := parsePath(suffixPath, allowMissingKeys)
	if err != nil {
		return nil, err
	}
	subResults, err := subJp.FindResults(item)
	if err != nil {
		if allowMissingKeys && strings.Contains(err.Error(), "array index out of bounds") {
			if raw, ok, recErr := evalMultiBranchRaw(ctx, suffixPath, item, allowMissingKeys); ok {
				return raw, recErr
			}
			return nil, nil
		}
		return nil, err
	}
	var raw []reflect.Value
	for _, row := range subResults {
		raw = append(raw, row...)
	}
	return raw, nil
}

func isBranchingNode(node jsonpath.Node) bool {
	switch n := node.(type) {
	case *jsonpath.WildcardNode:
		return true
	case *jsonpath.RecursiveNode:
		return true
	case *jsonpath.UnionNode:
		return true
	case *jsonpath.FilterNode:
		return true
	case *jsonpath.ArrayNode:
		return !n.Params[1].Derived
	default:
		return false
	}
}

func formatNodes(nodes []jsonpath.Node) (string, error) {
	var sb strings.Builder
	for i, n := range nodes {
		s, err := formatNode(n)
		if err != nil {
			return "", err
		}
		if i > 0 {
			if _, ok := nodes[i-1].(*jsonpath.RecursiveNode); ok {
				s = strings.TrimPrefix(s, ".")
			}
		}
		sb.WriteString(s)
	}
	return sb.String(), nil
}

func formatNode(node jsonpath.Node) (string, error) {
	switch n := node.(type) {
	case *jsonpath.FieldNode:
		return "." + escapeFieldName(n.Value), nil
	case *jsonpath.WildcardNode:
		return ".*", nil
	case *jsonpath.RecursiveNode:
		return "..", nil
	case *jsonpath.ArrayNode:
		return formatArrayNode(n), nil
	case *jsonpath.UnionNode:
		var parts []string
		for _, listNode := range n.Nodes {
			elem, err := formatUnionElement(listNode.Nodes)
			if err != nil {
				return "", err
			}
			parts = append(parts, elem)
		}
		return "[" + strings.Join(parts, ",") + "]", nil
	case *jsonpath.FilterNode:
		left, err := formatNodes(n.Left.Nodes)
		if err != nil {
			return "", err
		}
		if n.Operator == "exists" {
			return fmt.Sprintf("[?(%s)]", left), nil
		}
		right, err := formatNodes(n.Right.Nodes)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("[?(%s %s %s)]", left, n.Operator, right), nil
	case *jsonpath.TextNode:
		return strconv.Quote(n.Text), nil
	case *jsonpath.IntNode:
		return strconv.Itoa(n.Value), nil
	case *jsonpath.FloatNode:
		return strconv.FormatFloat(n.Value, 'f', -1, 64), nil
	case *jsonpath.BoolNode:
		return strconv.FormatBool(n.Value), nil
	case *jsonpath.IdentifierNode:
		return n.Name, nil
	default:
		return "", fmt.Errorf("unsupported JSONPath node %v", node)
	}
}

func escapeFieldName(name string) string {
	var sb strings.Builder
	for _, r := range name {
		switch r {
		case '\\', '.', ',', '[', ']', '$', '@', '{', '}', ' ', '\t', '\r', '\n', '*':
			sb.WriteRune('\\')
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func formatArrayNode(n *jsonpath.ArrayNode) string {
	if n.Params[1].Derived {
		return fmt.Sprintf("[%d]", n.Params[0].Value)
	}
	if !n.Params[0].Known && !n.Params[1].Known && !n.Params[2].Known {
		return "[*]"
	}
	var start, end, step string
	if n.Params[0].Known {
		start = strconv.Itoa(n.Params[0].Value)
	}
	if n.Params[1].Known {
		end = strconv.Itoa(n.Params[1].Value)
	}
	if n.Params[2].Known {
		step = ":" + strconv.Itoa(n.Params[2].Value)
	}
	return fmt.Sprintf("[%s:%s%s]", start, end, step)
}

func formatUnionElement(nodes []jsonpath.Node) (string, error) {
	if len(nodes) == 1 {
		switch n := nodes[0].(type) {
		case *jsonpath.FieldNode:
			return "'" + escapeFieldName(n.Value) + "'", nil
		case *jsonpath.ArrayNode:
			arrStr := formatArrayNode(n)
			return strings.TrimSuffix(strings.TrimPrefix(arrStr, "["), "]"), nil
		case *jsonpath.IntNode:
			return strconv.Itoa(n.Value), nil
		}
	}
	return formatNodes(nodes)
}

func collectResults(results [][]reflect.Value) ([]any, error) {
	a := make([]any, 0)
	for _, r := range results {
		for _, v := range r {
			o, err := extractResult(v)
			if err != nil {
				return nil, err
			}
			if o != nil {
				a = append(a, o)
			}
		}
	}
	return a, nil
}

func parsePath(p string, allowMissingKeys bool) (*jsonpath.JSONPath, error) {
	jp := jsonpath.New("<jsonpath>")
	err := jp.Parse("{" + p + "}")
	if err != nil {
		return nil, err
	}
	jp.AllowMissingKeys(allowMissingKeys)
	return jp, nil
}

func extractResult(v reflect.Value) (any, error) {
	// a nil input gives an invalid (zero) value
	if !v.IsValid() {
		return nil, nil
	}

	if v.CanInterface() {
		return v.Interface(), nil
	}

	return nil, fmt.Errorf("JSONPath couldn't access field")
}
