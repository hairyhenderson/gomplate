package gomplate

import (
	"context"
	"sync"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CELTracker", func() {
	newTrackedContext := func(tracker *CELTracker) commonsContext.Context {
		return WithCELTracker(commonsContext.NewContext(context.Background()), tracker)
	}

	It("tracks native evaluation state only when requested", func() {
		tracker := NewCELTracker()
		template := Template{Expression: "a == b && c > d"}
		values := map[string]any{"a": 1, "b": 2, "c": 3, "d": 4}

		result, err := RunExpression(values, template)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(false))
		Expect(tracker.Snapshot()).To(Equal(CELTraceSnapshot{}))

		result, err = RunExpressionContext(newTrackedContext(tracker), values, template)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(false))

		snapshot := tracker.Snapshot()
		Expect(snapshot.AST).NotTo(BeNil())
		Expect(snapshot.Details).NotTo(BeNil())
		Expect(snapshot.Details.State().IDs()).NotTo(BeEmpty())
		Expect(snapshot.Output.Value()).To(Equal(false))
	})

	It("retains partial native state when evaluation returns an error", func() {
		tracker := NewCELTracker()
		expression := "numerator / denominator == expected"

		_, err := RunExpressionContext(newTrackedContext(tracker), map[string]any{
			"numerator":   8,
			"denominator": 0,
			"expected":    2,
		}, Template{Expression: expression})
		Expect(err).To(HaveOccurred())

		snapshot := tracker.Snapshot()
		Expect(snapshot.AST).NotTo(BeNil())
		Expect(snapshot.AST.Source().Content()).To(Equal(expression))
		Expect(snapshot.Details).NotTo(BeNil())
		Expect(snapshot.Details.State().IDs()).NotTo(BeEmpty())
		Expect(snapshot.Output).NotTo(BeNil())
	})

	It("preserves multiline source locations", func() {
		tracker := NewCELTracker()
		expression := "a == b &&\n  c > d"

		_, err := RunExpressionContext(newTrackedContext(tracker), map[string]any{
			"a": 1,
			"b": 1,
			"c": 3,
			"d": 4,
		}, Template{Expression: expression})
		Expect(err).NotTo(HaveOccurred())

		snapshot := tracker.Snapshot()
		Expect(snapshot.AST.Source().Content()).To(Equal(expression))
		ids := snapshot.Details.State().IDs()
		valueLines := make([]int, 0, len(ids))
		for _, id := range ids {
			valueLines = append(valueLines, snapshot.AST.NativeRep().SourceInfo().GetStartLocation(id).Line())
		}
		Expect(valueLines).To(ContainElement(2))
	})

	It("rejects concurrent reuse and can be reused after evaluation", func() {
		tracker := NewCELTracker()
		started := make(chan struct{})
		release := make(chan struct{})
		var startOnce sync.Once
		blockingFunction := cel.Function("block",
			cel.Overload("block_int", []*cel.Type{cel.IntType}, cel.IntType,
				cel.UnaryBinding(func(value ref.Val) ref.Val {
					startOnce.Do(func() { close(started) })
					<-release
					return value
				})))
		template := Template{Expression: "block(value) == 1", CelEnvs: []cel.EnvOption{blockingFunction}}
		firstDone := make(chan error, 1)

		go func() {
			_, err := RunExpressionContext(newTrackedContext(tracker), map[string]any{"value": 1}, template)
			firstDone <- err
		}()
		<-started

		_, err := RunExpressionContext(newTrackedContext(tracker), map[string]any{"value": 1}, template)
		Expect(err).To(MatchError(ContainSubstring("CEL tracker is already in use")))
		close(release)
		Expect(<-firstDone).NotTo(HaveOccurred())

		_, err = RunExpressionContext(newTrackedContext(tracker), map[string]any{"value": 1}, template)
		Expect(err).NotTo(HaveOccurred())
	})

	It("leaves the normal program cache unchanged", func() {
		celExpressionCache.Flush()
		template := Template{Expression: "value == 1"}
		values := map[string]any{"value": 1}

		_, err := RunExpression(values, template)
		Expect(err).NotTo(HaveOccurred())
		Expect(celExpressionCache.ItemCount()).To(Equal(1))

		tracker := NewCELTracker()
		_, err = RunExpressionContext(newTrackedContext(tracker), values, template)
		Expect(err).NotTo(HaveOccurred())
		Expect(celExpressionCache.ItemCount()).To(Equal(1))
	})

	It("does not publish a snapshot for compile errors", func() {
		tracker := NewCELTracker()
		_, err := RunExpressionContext(newTrackedContext(tracker), nil, Template{Expression: "missing("})
		Expect(err).To(HaveOccurred())
		Expect(tracker.Snapshot()).To(Equal(CELTraceSnapshot{}))
	})
})
