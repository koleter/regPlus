package regexp

import (
	"container/list"
	"fmt"
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

// 测试正则的stringVar的匹配次数限制功能
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
			// 共5处位置,5个字符串在随机位置出现,可以出现也可以不出现,出现的总次数为3次
			"No.6", "a${word}?b${word}?c${word}?d${word}?e${word}?", "aaqwbchellodaadeapple", "word", 3, 3, []string{"hello", "hallo", "world", "aad", "aqw"}, "aaqwbchellodaade", "",
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
	mustCompile := MustCompile("\\(${word}\\)")
	mustCompile.RegisterStringVar("word", []string{"hello", "hallo", "world", "aad", "aqw"}...)
	submatch := mustCompile.FindAllString("a(aad)b(aad)cd(hello)fg", -1)
	assert.Equal(t, submatch, []string{"(aad)", "(aad)", "(hello)"})
}

func TestRegexp_RegisterRegVar(t *testing.T) {
	cases := []struct {
		name   string
		reg    string
		text   []string
		word   string
		regs   []*Regexp
		expect []string
	}{
		{
			"No.1", "a@{var}b@{var}", []string{"a302bacR"}, "var", []*Regexp{MustCompile("[a-z]*"), MustCompile("\\d+")}, []string{""},
		},
	}

	for _, tt := range cases {
		mustCompile := MustCompile(tt.reg)
		mustCompile.RegisterRegVar(tt.word, tt.regs...)
		for i, str := range tt.text {
			t.Run(fmt.Sprintf("%s#%d", tt.name, i), func(t *testing.T) {
				findString := mustCompile.FindString(str)
				assert.Equal(t, findString, tt.expect[i])
			})
		}
	}
}

func TestOnePass(t *testing.T) {
	mustCompile := MustCompile("asdfg")
	findString := mustCompile.FindString("asdfe")
	fmt.Println(findString)
}

func TestList(t *testing.T) {
	l := list.List{}
	back := l.PushBack(3)
	l.InsertBefore(1, back)
	for node := l.Front(); node != nil; node = node.Next() {
		val := node.Value
		fmt.Println(val)
	}
}

func TestReuseReg(t *testing.T) {
	mustCompile := MustCompile(".*(\\d+)")

	findString := mustCompile.FindAllStringSubmatch("w2012", -1)
	fmt.Println(findString)

	findString = mustCompile.FindAllStringSubmatch("w2012", -1)
	fmt.Println(findString)
}
