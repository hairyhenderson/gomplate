package cli

import (
	"testing"
)

func TestUTokenize(t *testing.T) {
	cases := []struct {
		usage    string
		expected []*uToken
	}{
		{"OPTIONS", []*uToken{{utOptions, "OPTIONS", 0}}},

		{"XOPTIONS", []*uToken{{utPos, "XOPTIONS", 0}}},
		{"OPTIONSX", []*uToken{{utPos, "OPTIONSX", 0}}},
		{"ARG", []*uToken{{utPos, "ARG", 0}}},
		{"ARG42", []*uToken{{utPos, "ARG42", 0}}},
		{"ARG_EXTRA", []*uToken{{utPos, "ARG_EXTRA", 0}}},

		{"ARG1 ARG2", []*uToken{{utPos, "ARG1", 0}, {utPos, "ARG2", 5}}},
		{"ARG1  ARG2", []*uToken{{utPos, "ARG1", 0}, {utPos, "ARG2", 6}}},

		{"[ARG]", []*uToken{{utOpenSq, "[", 0}, {utPos, "ARG", 1}, {utCloseSq, "]", 4}}},
		{"[ ARG ]", []*uToken{{utOpenSq, "[", 0}, {utPos, "ARG", 2}, {utCloseSq, "]", 6}}},
		{"ARG [ARG2 ]", []*uToken{{utPos, "ARG", 0}, {utOpenSq, "[", 4}, {utPos, "ARG2", 5}, {utCloseSq, "]", 10}}},
		{"ARG [ ARG2]", []*uToken{{utPos, "ARG", 0}, {utOpenSq, "[", 4}, {utPos, "ARG2", 6}, {utCloseSq, "]", 10}}},

		{"...", []*uToken{{utRep, "...", 0}}},
		{"ARG...", []*uToken{{utPos, "ARG", 0}, {utRep, "...", 3}}},
		{"ARG ...", []*uToken{{utPos, "ARG", 0}, {utRep, "...", 4}}},
		{"[ARG...]", []*uToken{{utOpenSq, "[", 0}, {utPos, "ARG", 1}, {utRep, "...", 4}, {utCloseSq, "]", 7}}},

		{"|", []*uToken{{utChoice, "|", 0}}},
		{"ARG|ARG2", []*uToken{{utPos, "ARG", 0}, {utChoice, "|", 3}, {utPos, "ARG2", 4}}},
		{"ARG |ARG2", []*uToken{{utPos, "ARG", 0}, {utChoice, "|", 4}, {utPos, "ARG2", 5}}},
		{"ARG| ARG2", []*uToken{{utPos, "ARG", 0}, {utChoice, "|", 3}, {utPos, "ARG2", 5}}},

		{"[OPTIONS]", []*uToken{{utOpenSq, "[", 0}, {utOptions, "OPTIONS", 1}, {utCloseSq, "]", 8}}},

		{"-p", []*uToken{{utShortOpt, "-p", 0}}},
		{"-X", []*uToken{{utShortOpt, "-X", 0}}},

		{"--force", []*uToken{{utLongOpt, "--force", 0}}},
		{"--sig-proxy", []*uToken{{utLongOpt, "--sig-proxy", 0}}},

		{"-aBc", []*uToken{{utOptSeq, "aBc", 1}}},
		{"--", []*uToken{{utDoubleDash, "--", 0}}},
		{"=<bla>", []*uToken{{utOptValue, "=<bla>", 0}}},
		{"=<bla-bla>", []*uToken{{utOptValue, "=<bla-bla>", 0}}},
		{"=<bla--bla>", []*uToken{{utOptValue, "=<bla--bla>", 0}}},
		{"-p=<file-path>", []*uToken{{utShortOpt, "-p", 0}, {utOptValue, "=<file-path>", 2}}},
		{"--path=<absolute-path>", []*uToken{{utLongOpt, "--path", 0}, {utOptValue, "=<absolute-path>", 6}}},
	}
	for _, c := range cases {
		t.Logf("test %s", c.usage)
		tks, err := uTokenize(c.usage)
		if err != nil {
			t.Errorf("[Tokenize '%s']: Unexpected error: %v", c.usage, err)
			continue
		}

		t.Logf("actual: %v\n", tks)
		if len(tks) != len(c.expected) {
			t.Errorf("[Tokenize '%s']: token count mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, c.expected, tks)
			continue
		}

		for i, actual := range tks {
			expected := c.expected[i]
			switch {
			case actual.typ != expected.typ:
				t.Errorf("[Tokenize '%s']: token type mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, expected, actual)
			case actual.val != expected.val:
				t.Errorf("[Tokenize '%s']: token text mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, expected, actual)
			case actual.pos != expected.pos:
				t.Errorf("[Tokenize '%s']: token pos mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, expected, actual)
			}
		}

	}
}

func TestUTokenizeErrors(t *testing.T) {
	cases := []struct {
		usage string
		pos   int
	}{
		{"-", 1},
		{"---x", 2},
		{"-x-", 2},

		{"=", 1},
		{"=<", 2},
		{"=<dsdf", 6},
		{"=<>", 2},
	}

	for _, c := range cases {
		t.Logf("test %s", c.usage)
		tks, err := uTokenize(c.usage)
		if err == nil {
			t.Errorf("Tokenize('%s') should have failed, instead got %v", c.usage, tks)
			continue
		}
		t.Logf("Got expected error %v", err)
		if err.pos != c.pos {
			t.Errorf("[Tokenize '%s']: error pos mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, c.pos, err.pos)

		}
	}
}
