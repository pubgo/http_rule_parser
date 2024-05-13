package http_rule

import (
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/pretty"
)

const url1 = "/v1/users/{aa.ba.ca=hh/*}/hello/{hello.abc}/*/hhh/*/messages/{messageId=nn/ss/**}:change"

func TestMatch(t *testing.T) {
	route := ParseToRoute(assert.Exit1(Parse(url1)))
	vars := assert.Must1(route.Match([]string{"v1", "users", "hh", "123", "hello", "vvv", "*", "hhh", "*", "messages", "nn", "ss", "vv", "33"}, "change"))
	t.Log(vars)
}

func TestParse(t *testing.T) {
	ini := assert.Must1(parser.Parse("",
		strings.NewReader(url1),
		participle.Trace(os.Stdout),
	))

	pretty.Println(ini)
	pp := ParseToRoute(ini)
	pretty.Println(pp)

	t.Log(pp.String())
	p1 := assert.Must1(parser.Parse("", strings.NewReader(pp.String())))
	if pp2 := ParseToRoute(p1).String(); pp2 != pp.String() {
		t.Fatal("not equal", pp2, pp.String())
	}
}

func BenchmarkParse(b *testing.B) {
	r := strings.NewReader(url1)
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse("", r)
	}
}
