package strings

import "testing"

func TestByteSizeFormating(t *testing.T) {
	tests := []struct {
		Bytes     uint64
		Humanized string
	}{
		{0, "0B"},
		{10, "10B"},
		{1023, "1023B"},
		{1024, "1K"},
		{1024 * 1024, "1M"},
		{1024*1024 + 53, "1M"},
		{5 * 1024 * 1024, "5M"},
		{1 * 1024 * 1024 * 1024, "1G"},
		{1 * 1024 * 1024 * 1024 * 1024, "1T"},
		{10 * 1024 * 1024 * 1024 * 1024, "10T"},
		{10 * 1024 * 1024 * 1024 * 1024 * 1024, "10P"},
		{10 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024, "10E"},
	}
	for _, tc := range tests {
		if HumanBytes(tc.Bytes) != tc.Humanized {
			t.Errorf("Failed for test case %v ", tc)
		}
	}
}
