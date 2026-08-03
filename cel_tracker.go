package gomplate

import (
	"fmt"
	"sync"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
)

type celTrackerContextKey struct{}

// CELTraceSnapshot exposes cel-go's native syntax and evaluation state for one
// completed evaluation. It is empty when compilation fails.
type CELTraceSnapshot struct {
	AST     *cel.Ast
	Details *cel.EvalDetails
	Output  ref.Val
}

// CELTracker captures native CEL evaluation state when attached to a context.
// A tracker may be reused sequentially, but not concurrently.
type CELTracker struct {
	mu       sync.Mutex
	active   bool
	snapshot CELTraceSnapshot
}

// NewCELTracker creates an opt-in native CEL evaluation tracker.
func NewCELTracker() *CELTracker {
	return &CELTracker{}
}

// WithCELTracker enables native CEL state tracking for evaluations using ctx.
func WithCELTracker(ctx commonsContext.Context, tracker *CELTracker) commonsContext.Context {
	return ctx.WithValue(celTrackerContextKey{}, tracker)
}

func (t *CELTracker) Snapshot() CELTraceSnapshot {
	if t == nil {
		return CELTraceSnapshot{}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.snapshot
}

func celTrackerFromContext(ctx commonsContext.Context) *CELTracker {
	tracker, _ := ctx.Value(celTrackerContextKey{}).(*CELTracker)
	return tracker
}

func (t *CELTracker) begin() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.active {
		return fmt.Errorf("CEL tracker is already in use")
	}
	t.active = true
	t.snapshot = CELTraceSnapshot{}
	return nil
}

func (t *CELTracker) abort() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.active = false
}

func (t *CELTracker) complete(ast *cel.Ast, details *cel.EvalDetails, output ref.Val) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.snapshot = CELTraceSnapshot{AST: ast, Details: details, Output: output}
}
