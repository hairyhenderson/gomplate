package gomplate

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/google/cel-go/common/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type registeredCELPerson struct {
	DisplayName string `json:"display_name"`
	Nickname    string `json:",omitempty"`
	Ignored     string `json:"-"`
}

type cachedCELPerson struct {
	Name string `json:"name"`
}

type describedCELPerson struct {
	Name string `json:"name"`
}

type concurrentCELPerson struct {
	DisplayName string `json:"display_name"`
}

type reflectedTypeCELPerson struct {
	Name string `json:"name"`
}

type reflectedValueCELPerson struct {
	Name string `json:"name"`
}

var _ = Describe("CEL native types", Ordered, func() {
	It("passes registered top-level values to CEL without serializing them", func() {
		person := registeredCELPerson{
			DisplayName: "Ada Lovelace",
			Nickname:    "Ada",
			Ignored:     "private",
		}
		Expect(RegisterType(person)).To(Succeed())

		result, err := RunExpression(map[string]any{"person": person}, Template{
			Expression: "person",
			CacheKey:   "cel-native-person",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(person))

		field, err := RunExpression(map[string]any{"person": &person}, Template{
			Expression: `person.display_name + ":" + person.Nickname`,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(field).To(Equal("Ada Lovelace:Ada"))

		_, err = RunExpression(map[string]any{"person": person}, Template{Expression: "person.Ignored"})
		Expect(err).To(MatchError(ContainSubstring("no such field: Ignored")))
	})

	It("invalidates cached programs when a native type is registered", func() {
		person := cachedCELPerson{Name: "Grace Hopper"}
		template := Template{Expression: "person", CacheKey: "cel-native-generation"}

		before, err := RunExpression(map[string]any{"person": person}, template)
		Expect(err).NotTo(HaveOccurred())
		Expect(before).To(Equal(map[string]any{"name": "Grace Hopper"}))

		Expect(RegisterType(person)).To(Succeed())

		after, err := RunExpression(map[string]any{"person": person}, template)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(person))
	})

	It("accepts self-describing native CEL types", func() {
		descriptor, err := types.NewNativeType(
			reflect.TypeOf(describedCELPerson{}),
			types.ParseStructField(jsonCELFieldName),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(RegisterType(descriptor)).To(Succeed())

		person := describedCELPerson{Name: "Katherine Johnson"}
		result, err := RunExpression(map[string]any{"person": person}, Template{Expression: "person"})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(person))
	})

	It("accepts reflect.Type and reflect.Value registrations", func() {
		Expect(RegisterType(reflect.TypeOf(reflectedTypeCELPerson{}))).To(Succeed())
		Expect(RegisterType(reflect.ValueOf(reflectedValueCELPerson{}))).To(Succeed())

		result, err := RunExpression(map[string]any{
			"typed":  reflectedTypeCELPerson{Name: "Mary Jackson"},
			"valued": reflectedValueCELPerson{Name: "Christine Darden"},
		}, Template{Expression: `typed.name + ":" + valued.name`})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("Mary Jackson:Christine Darden"))
	})

	It("rejects invalid types without publishing a new generation", func() {
		generation := currentNativeTypes().generation

		err := RegisterType(42)

		Expect(err).To(MatchError(ContainSubstring("unsupported reflect.Type")))
		Expect(currentNativeTypes().generation).To(Equal(generation))
	})

	It("treats repeated registrations as idempotent", func() {
		person := registeredCELPerson{}
		Expect(RegisterType(person)).To(Succeed())
		generation := currentNativeTypes().generation

		Expect(RegisterType(&person)).To(Succeed())

		Expect(currentNativeTypes().generation).To(Equal(generation))
	})

	It("supports concurrent registration snapshots and evaluation", func() {
		person := concurrentCELPerson{DisplayName: "Dorothy Vaughan"}
		errors := make(chan error, 16)
		var waitGroup sync.WaitGroup
		for range 8 {
			waitGroup.Add(2)
			go func() {
				defer waitGroup.Done()
				errors <- RegisterType(person)
			}()
			go func() {
				defer waitGroup.Done()
				_, err := RunExpression(map[string]any{"person": person}, Template{Expression: "person.display_name"})
				errors <- err
			}()
		}
		waitGroup.Wait()
		close(errors)

		for err := range errors {
			Expect(err).NotTo(HaveOccurred())
		}
	})
})

var _ = Describe("CEL regex program limits", func() {
	const limit = 10_000

	It("rejects oversized literal regex programs during compilation", func() {
		pattern := strings.Repeat("a?", limit+1)
		expression := fmt.Sprintf(`"a".matches(%q)`, pattern)

		_, err := RunExpression(nil, Template{Expression: expression})

		Expect(err).To(MatchError(ContainSubstring("regex program size")))
		Expect(err).To(MatchError(ContainSubstring("exceeds limit of 10000")))
	})

	It("rejects oversized dynamic regex programs during evaluation", func() {
		pattern := strings.Repeat("a?", limit+1)

		_, err := RunExpression(map[string]any{"pattern": pattern}, Template{
			Expression: `"a".matches(pattern)`,
		})

		Expect(err).To(MatchError(ContainSubstring("regex program size")))
		Expect(err).To(MatchError(ContainSubstring("exceeds limit of 10000")))
	})
})
