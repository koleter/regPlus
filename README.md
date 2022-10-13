# regPlus
Add the following functions on the basis of the official go package regexp. For usage, please refer to the test cases in *regexp_test.go*
* ## string variable
In regular expressions, a string variable can be marked with a sequence of characters 
like *${word}* .You can then register some strings for this string variable. This string 
variable will match one of these strings when a regular expression is matching.These strings can 
only be matched once.
<div style="height: 1em"></div>
For example:

```go
Compile := MustCompile("a${word}b${word}c")
Compile.RegisterStringVar("word", "abc", "def")
```
This regular expression can match string like *aabcbdefc* (The first ${word} matches 
abc and the second ${word} matches def) or *adefbabcc* (The first ${word} matches
def and the second ${word} matches abc).
It don't match *aabcbabcc* because there is only one "abc" is registerd on string variable.
<div style="height: 1em"></div>

You can limit the number of times string variable is used.
For example:
```go
Compile := MustCompile("${word}*")
Compile.RegisterStringVar("word", []string{"a", "b", "c", "d", "e"}...)
// set ${word} can match 2 to 4 times
Compile.SetStringVarLimit("word", 2, 4)
```
This regular expression can match "ab", "bc", "ea", "bce", "deac", "dcab", etc.

* ## reg variable
In regular expressions, you can use function RegisterRegVar to register a reg variable. Reg variable can be marked with a sequence of characters
like *@{word}* . It is used in a similar way to string variable. It is not replaced by 
string, but the regular expressions.

For example:
```go
mustCompile := MustCompile("a@{var}b@{var}")
mustCompile.RegisterRegVar("var", []*Regexp{MustCompile("\\d+"), MustCompile("[a-z]*")}...)
```
