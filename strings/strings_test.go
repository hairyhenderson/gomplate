package strings

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndent(t *testing.T) {
	actual := "hello\nworld\n!"
	expected := "  hello\n  world\n  !"
	assert.Equal(t, actual, Indent(0, "  ", actual))
	assert.Equal(t, actual, Indent(-1, "  ", actual))
	assert.Equal(t, expected, Indent(1, "  ", actual))
	assert.Equal(t, "\n", Indent(1, "  ", "\n"))
	assert.Equal(t, "  foo\n", Indent(1, "  ", "foo\n"))
	assert.Equal(t, "   foo", Indent(1, "   ", "foo"))
	assert.Equal(t, "   foo", Indent(3, " ", "foo"))

	// indenting with newline is not permitted
	assert.Equal(t, "foo", Indent(3, "\n", "foo"))
}

func TestTrunc(t *testing.T) {
	assert.Equal(t, "", Trunc(5, ""))
	assert.Equal(t, "", Trunc(0, "hello, world"))
	assert.Equal(t, "hello", Trunc(5, "hello, world"))
	assert.Equal(t, "hello, world", Trunc(12, "hello, world"))
	assert.Equal(t, "hello, world", Trunc(42, "hello, world"))
	assert.Equal(t, "hello, world", Trunc(-1, "hello, world"))
}

func TestShellQuote(t *testing.T) {
	assert.Equal(t, `''`, ShellQuote(``))
	assert.Equal(t, `'foo'`, ShellQuote(`foo`))
	assert.Equal(t, `'hello "world"'`, ShellQuote(`hello "world"`))
	assert.Equal(t, `'it'"'"'s its'`, ShellQuote(`it's its`))
}

func TestSort(t *testing.T) {
	in := []string{}
	expected := []string{}
	assert.EqualValues(t, expected, Sort(in))

	in = []string{"c", "a", "b"}
	expected = []string{"a", "b", "c"}
	assert.EqualValues(t, expected, Sort(in))

	in = []string{"42", "45", "18"}
	expected = []string{"18", "42", "45"}
	assert.EqualValues(t, expected, Sort(in))
}

func TestCaseFuncs(t *testing.T) {
	testdata := []struct{ in, s, k, c string }{
		{"  Foo bar ", "Foo_bar", "Foo-bar", "FooBar"},
		{"foo  bar", "foo_bar", "foo-bar", "fooBar"},
		{" baz\tqux  ", "baz_qux", "baz-qux", "bazQux"},
		{"Hello, World!", "Hello_world", "Hello-world", "HelloWorld"},
		{"grüne | Straße", "grüne_straße", "grüne-straße", "grüneStraße"},
	}
	for _, d := range testdata {
		assert.Equal(t, d.s, SnakeCase(d.in))
		assert.Equal(t, d.k, KebabCase(d.in))
		assert.Equal(t, d.c, CamelCase(d.in))
	}
}

func TestWordWrap(t *testing.T) {
	in := "a short line that needs no wrapping"
	assert.Equal(t, in, WordWrap(in, WordWrapOpts{Width: 40}))

	in = "a short line that needs wrapping"
	out := `a short
line that
needs
wrapping`

	assert.Equal(t, out, WordWrap(in, WordWrapOpts{Width: 9}))
	in = "a short line that needs wrapping"
	out = `a short \
line that \
needs \
wrapping`
	assert.Equal(t, out, WordWrap(in, WordWrapOpts{Width: 9, LBSeq: " \\\n"}))

	out = `There shouldn't be any wrapping of long words or URLs because that would break
things very badly. To wit:
https://example.com/a/super-long/url/that-shouldnt-be?wrapped=for+fear+of#the-breaking-of-functionality
should appear on its own line, regardless of the desired word-wrapping width
that has been set.`
	in = strings.ReplaceAll(out, "\n", " ")
	assert.Equal(t, out, WordWrap(in, WordWrapOpts{}))

	// TODO: get these working - need to switch to a word-wrapping package that
	// can handle multi-byte characters!
	//
	// 	out = `ΤΟΙΣ πᾶσι χρόνος καὶ καιρὸς τῷ παντὶ πράγματι ὑπὸ τὸν οὐρανόν. καιρὸς τοῦ
	// τεκεῖν καὶ καιρὸς τοῦ ἀποθανεῖν, καιρὸς τοῦ φυτεῦσαι καὶ καιρὸς τοῦ ἐκτῖλαι τὸ
	// πεφυτευμένον, καιρὸς τοῦ ἀποκτεῖναι καὶ καιρὸς τοῦ ἰάσασθαι, καιρὸς τοῦ
	// καθελεῖν καὶ καιρὸς τοῦ οἰκοδομεῖν, καιρὸς τοῦ κλαῦσαι καὶ καιρὸς τοῦ γελάσαι,
	// καιρὸς τοῦ κόψασθαι καὶ καιρὸς τοῦ ὀρχήσασθαι, καιρὸς τοῦ βαλεῖν λίθους καὶ
	// καιρὸς τοῦ συναγαγεῖν λίθους, καιρὸς τοῦ περιλαβεῖν καὶ καιρὸς τοῦ μακρυνθῆναι
	// ἀπὸ περιλήψεως, καιρὸς τοῦ ζητῆσαι καὶ καιρὸς τοῦ ἀπολέσαι, καιρὸς τοῦ φυλάξαι
	// καὶ καιρὸς τοῦ ἐκβαλεῖν, καιρὸς τοῦ ρῆξαι καὶ καιρὸς τοῦ ράψαι, καιρὸς τοῦ
	// σιγᾶν καὶ καιρὸς τοῦ λαλεῖν, καιρὸς τοῦ φιλῆσαι καὶ καιρὸς τοῦ μισῆσαι, καιρὸς
	// πολέμου καὶ καιρὸς εἰρήνης.`
	// 	in = strings.ReplaceAll(out, "\n", " ")
	// 	assert.Equal(t, out, WordWrap(in, WordWrapOpts{}))

	// TODO: get these working - need to switch to a word-wrapping package that
	// understands multi-byte and correctly identifies line-breaking opportunities
	// for non-latin languages.
	//
	// 	out = `何事にも定まった時があります。
	// 生まれる時、死ぬ時、植える時、収穫の時、
	// 殺す時、病気が治る時、壊す時、やり直す時、
	// 泣く時、笑う時、悲しむ時、踊る時、
	// 石をばらまく時、石をかき集める時、
	// 抱きしめる時、抱きしめない時、
	// 何かを見つける時、物を失う時、
	// 大切にしまっておく時、遠くに投げ捨てる時、
	// 引き裂く時、修理する時、黙っている時、口を開く時、
	// 愛する時、憎む時、戦う時、和解する時。`
	// 	in = strings.ReplaceAll(out, "\n", " ")
	// 	assert.Equal(t, out, WordWrap(in, WordWrapOpts{Width: 100}))
}

func TestSkipLines(t *testing.T) {
	out, _ := SkipLines(2, "\nfoo\nbar\n\nbaz")
	assert.Equal(t, "bar\n\nbaz", out)

	out, _ = SkipLines(0, "foo\nbar\n\nbaz")
	assert.Equal(t, "foo\nbar\n\nbaz", out)

	_, err := SkipLines(-1, "foo\nbar\n\nbaz")
	assert.Error(t, err)

	out, err = SkipLines(4, "foo\nbar\n\nbaz")
	require.NoError(t, err)
	assert.Equal(t, "", out)
}
