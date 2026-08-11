package gomplate

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type benchAddress struct {
	City    string `json:"city_name"`
	Country string `json:"country,omitempty"`
}

type benchPerson struct {
	Name      string         `json:"name"`
	Age       int            `json:"age,omitempty"`
	ID        uuid.UUID      `json:"id"`
	Duration  time.Duration  `json:"duration"`
	Address   *benchAddress  `json:",omitempty"`
	MetaData  map[string]any `json:",omitempty"`
	Codes     []string       `json:",omitempty"`
	Addresses []benchAddress `json:"addresses,omitempty"`
}

func BenchmarkSerialize(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("items=%d/native_values=true", size), func(b *testing.B) {
			input := newSerializeBenchmarkInput(size)
			b.ReportAllocs()
			for b.Loop() {
				if _, err := Serialize(input); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSerialize_NoNativeTypes(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("items=%d/native_values=false", size), func(b *testing.B) {
			input := newPlainBenchmarkInput(size)
			b.ReportAllocs()
			for b.Loop() {
				if _, err := Serialize(input); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func newSerializeBenchmarkInput(size int) map[string]any {
	items := make([]any, size)
	identifier := uuid.MustParse("0192f0a4-1234-7000-8000-aaaaaaaaaaaa")
	for i := range items {
		items[i] = benchPerson{
			Name:     fmt.Sprintf("person-%d", i),
			Age:      i % 100,
			ID:       identifier,
			Duration: time.Duration(i) * time.Millisecond,
			Address:  &benchAddress{City: "Kathmandu", Country: "Nepal"},
			MetaData: map[string]any{
				"index": i, "enabled": i%2 == 0, "uuid": identifier,
				"duration": time.Duration(i) * time.Second,
			},
			Codes: []string{"GO", "JS", "CEL"},
			Addresses: []benchAddress{
				{City: "Kathmandu", Country: "Nepal"},
				{City: "Lalitpur", Country: "Nepal"},
				{City: "Bhaktapur", Country: "Nepal"},
			},
		}
	}
	return map[string]any{
		"id":       identifier,
		"started":  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		"duration": 5 * time.Minute,
		"items":    items,
		"nested": map[string]any{
			"bytes":     []byte("hello world"),
			"labels":    map[string]string{"app": "gomplate", "bench": "serialize"},
			"durations": []time.Duration{time.Second, time.Minute, time.Hour},
		},
	}
}

func newPlainBenchmarkInput(size int) map[string]any {
	items := make([]any, size)
	for i := range items {
		items[i] = map[string]any{
			"name":    fmt.Sprintf("person-%d", i),
			"age":     i % 100,
			"enabled": i%2 == 0,
			"codes":   []string{"GO", "JS", "CEL"},
			"address": map[string]any{"city": "Kathmandu", "country": "Nepal"},
		}
	}
	return map[string]any{
		"started": "2026-01-01T00:00:00Z",
		"items":   items,
		"nested": map[string]any{
			"labels": map[string]string{"app": "gomplate", "bench": "serialize"},
		},
	}
}
