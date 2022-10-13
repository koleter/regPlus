package regPlus

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegexp_RegisterStringVar(t *testing.T) {
	cases := []struct {
		name    string
		reg     string
		text    string
		word    string
		strs    []string
		expect  string
		wantErr string
	}{
		{
			"No.1", "a\\(${word}\\)b\\(${word}\\)cd", "a(abc)b(def)cd e", "word", []string{"abc", "def"}, "a(abc)b(def)cd", "",
		},
		{
			"No.2", "a\\(${word}\\)b\\(${word}\\)cd", "a(def)b(abc)cd e", "word", []string{"abc", "def"}, "a(def)b(abc)cd", "",
		},
		{
			"No.3", "a\\(${word}\\)b\\(${word}\\)cd", "a(def)b(def)cd e", "word", []string{"abc", "def"}, "", "",
		},
		{
			"No.4", "a\\(${word}\\)b\\(${word}\\)c|a\\(${word}\\)b\\(${word}\\)cd", "a(def)b(abc)cd e", "word", []string{"abc", "def"}, "a(def)b(abc)c", "",
		},
		{
			"No.5", "a\\(${word}\\)b\\(${word}\\)cd", "a(abc)b(abcde)cd e", "word", []string{"abc", "dabc", "abcde"}, "a(abc)b(abcde)cd", "",
		},
		{
			"No.6", "a\\(${word}\\)b\\(${word}\\)cd", "a(abcde)b(abc)cd e", "word", []string{"abc", "dabc", "abcde"}, "a(abcde)b(abc)cd", "",
		},
		{
			"No.7", "a\\(${word}\\)b\\(${word}\\)cd\\(${word}\\)", "a(abcde)b(abc)cd(abc) e", "word", []string{"abc", "dabc", "abcde"}, "", "",
		},
		{
			"No.8", "a\\(${word}\\)b\\(${word}\\)cd", "a(abc)b(defg)cd e", "word", []string{"abc", "def"}, "", "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			Compile, err := Compile(tt.reg)
			if err != nil {
				assert.Equal(t, err.Error(), tt.expect)
			} else {
				Compile.RegisterStringVar(tt.word, tt.strs...)
				findString := Compile.FindString(tt.text)
				assert.Equal(t, findString, tt.expect)
			}
		})
	}
}

func TestRegexp_RegisterStringVar_WithLimit(t *testing.T) {
	cases := []struct {
		name     string
		reg      string
		text     string
		word     string
		min, max int
		strs     []string
		expect   string
		wantErr  string
	}{
		{
			"No.1", "a\\(${word}\\)b\\(${word}\\)cd", "a(h中u)b(world)cdfg", "word", 2, 3, []string{"hello", "h中u", "world", "aad", "aqw"}, "a(h中u)b(world)cd", "",
		},
		{
			"No.2", "a\\(${word}\\)b\\(${word}\\)cd", "a(aad)b(world)cdfg", "word", 3, 3, []string{"hello", "hallo", "world", "aad", "aqw"}, "", "",
		},
		{
			"No.3", "a\\(${word}\\)b\\(${word}\\)cd\\(${word}\\)", "a(aad)b(world)cd(hello)fg", "word", 1, 3, []string{"hello", "hallo", "world", "aad", "aqw"}, "a(aad)b(world)cd(hello)", "",
		},
		{
			"No.4", "a${word}?", "aaqw", "word", 1, 2, []string{"hello", "hallo", "world", "aad", "aqw"}, "aaqw", "",
		},
		{
			"No.5", "a${word}?", "aaqw", "word", 0, 0, []string{"hello", "hallo", "world", "aad", "aqw"}, "a", "",
		},
		{
			// 5 positions, 5 strings appear in random positions, may or may not appear, the total number of times the string appears is 3
			"No.6", "a${word}?b${word}?c${word}?d${word}?e${word}?", "aaqwbchellodaadeapple", "word", 3, 3, []string{"hello", "hallo", "world", "aad", "aqw"}, "aaqwbchellodaade", "",
		},
		{
			"No.7", "${word}*", "abdec", "word", 2, 4, []string{"a", "b", "c", "d", "e"}, "abde", "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			Compile, err := Compile(tt.reg)
			if err != nil {
				assert.Equal(t, err.Error(), tt.expect)
			} else {
				Compile.RegisterStringVar(tt.word, tt.strs...)
				Compile.SetStringVarLimit(tt.word, tt.min, tt.max)
				findString := Compile.FindString(tt.text)
				assert.Equal(t, findString, tt.expect)
			}
		})
	}
}

func TestRegexp_RegisterStringVar_FindAllString(t *testing.T) {
	cases := []struct {
		name   string
		text   string
		expect []string
	}{
		{
			"No.1", "a(aad)b(aad)cd(hello)fg", []string{"(aad)", "(aad)", "(hello)"},
		},
		{
			"No.2", "a(aad)b(hallo)cd(hello)fg", []string{"(aad)", "(hallo)", "(hello)"},
		},
	}

	mustCompile := MustCompile("\\(${word}\\)")
	mustCompile.RegisterStringVar("word", []string{"hello", "hallo", "world", "aad", "aqw"}...)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			submatch := mustCompile.FindAllString(tt.text, -1)
			assert.Equal(t, submatch, tt.expect)
		})
	}
}

func TestRegexp_RegisterRegVar(t *testing.T) {
	cases := []struct {
		name   string
		reg    string
		text   string
		word   string
		regs   []*Regexp
		expect string
	}{
		{
			"No.1", "a@{var}2b@{var}", "a302bacR", "var", []*Regexp{MustCompile("\\d+"), MustCompile("[a-z]*")}, "a302bac",
		},
		{
			"No.2", "a@{var}2b@{var}", "a302bacR", "var", []*Regexp{MustCompile("[a-z]*"), MustCompile("\\d+")}, "a302bac",
		},
		{
			"No.3", "a@{var}2b@{var}", "a502q302bacR", "var", []*Regexp{MustCompile("\\d+"), MustCompile("[a-z]*")}, "a502q302bac",
		},
		{
			"No.4", "a@{var}2b@{var}", "a502q302bacR", "var", []*Regexp{MustCompile("^\\d+"), MustCompile("[a-z]*")}, "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mustCompile := MustCompile(tt.reg)
			mustCompile.RegisterRegVar(tt.word, tt.regs...)
			findString := mustCompile.FindString(tt.text)
			assert.Equal(t, findString, tt.expect)
		})

	}
}

func TestRegexp_RegisterRegVarWithLimit(t *testing.T) {
	cases := []struct {
		name     string
		reg      string
		min, max int
		text     string
		word     string
		regs     []*Regexp
		expect   []string
	}{
		{
			"No.1", "a(@{var})2b(@{var})", 2, 3, "a302bacR", "var", []*Regexp{MustCompile("\\d+"), MustCompile("[a-z]*")}, []string{"a302bac", "30", "ac"},
		},
		{
			"No.2", "a(@{var})2b(@{var})", 3, 3, "a302bacR", "var", []*Regexp{MustCompile("\\d+"), MustCompile("[a-z]*")}, nil,
		},
		{
			"No.3", "a(@{var})2b(@{var})", 2, 3, "a302bacR", "var", []*Regexp{MustCompile("\\d+"), MustCompile("[a-z]*")}, []string{"a302bac", "30", "ac"},
		},
		{
			"No.4", "a(@{var})2b(@{var})", 2, 3, "a30002bacR", "var", []*Regexp{MustCompile("\\d+?"), MustCompile("[a-z]*")}, []string{"a30002bac", "3000", "ac"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mustCompile := MustCompile(tt.reg)
			mustCompile.RegisterRegVar(tt.word, tt.regs...)
			mustCompile.SetRegVarLimit(tt.word, tt.min, tt.max)
			findString := mustCompile.FindStringSubmatch(tt.text)
			assert.Equal(t, findString, tt.expect)
		})

	}
}

// Taking a SQL where condition as an example, it is required that num must set the upper and lower bounds of the query
func TestRegVarInSqlMatch(t *testing.T) {
	cases := []struct {
		name   string
		text   string
		expect string
	}{
		{
			"No.1", "where num >= 3", "",
		},
		{
			"No.2", "where num > 1 and num <= 6", "where num > 1 and num <= 6",
		},
		{
			"No.3", "where num > 4 and num < 9", "where num > 4 and num < 9",
		},
		{
			"No.4", "where 10 < num and num <= 40", "where 10 < num and num <= 40",
		},
		{
			"No.5", "where 10 < num and num > 40", "",
		},
	}
	mustCompile := MustCompile("where +@{board}.*@{board}.*")
	regexp := MustCompile("(@{lower}.*)+")
	regexp.RegisterRegVar("lower", MustCompile("num +<"), MustCompile("> +num"))

	r := MustCompile("(@{upper}.*)+")
	r.RegisterRegVar("upper", MustCompile("num +>"), MustCompile("< +num"))

	mustCompile.RegisterRegVar("board", regexp, r)

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			findString := mustCompile.FindString(tt.text)
			assert.Equal(t, findString, tt.expect)
		})
	}
}
