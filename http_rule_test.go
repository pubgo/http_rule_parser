package http_rule

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2"
	asserts "github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/pretty"
)

const url1 = "/v1/users/{aa.ba.ca=hh/*}/hello/{hello.abc}/*/hhh/*/messages/{messageId=nn/ss/**}:change"

func TestMatch(t *testing.T) {
	route := ParseToRoute(asserts.Exit1(Parse(url1)))
	vars := asserts.Must1(route.Match([]string{"v1", "users", "hh", "123", "hello", "vvv", "*", "hhh", "*", "messages", "nn", "ss", "vv", "33"}, "change"))
	asserts.If(len(vars) != 3, "not match")
	asserts.If(vars[0].Value != "hh/123", "not match")
	asserts.If(!reflect.DeepEqual(vars[0].Fields, []string{"aa", "ba", "ca"}), "not match")
	asserts.If(vars[1].Value != "vvv", "not match")
	asserts.If(!reflect.DeepEqual(vars[1].Fields, []string{"hello", "abc"}), "not match")
	asserts.If(vars[2].Value != "nn/ss/vv/33", "not match")
	asserts.If(!reflect.DeepEqual(vars[2].Fields, []string{"messageId"}), "not match")
	pretty.Println(vars)
}

func BenchmarkMatch(b *testing.B) {
	route := ParseToRoute(asserts.Exit1(Parse(url1)))
	for i := 0; i < b.N; i++ {
		_ = asserts.Must1(route.Match([]string{"v1", "users", "hh", "123", "hello", "vvv", "*", "hhh", "*", "messages", "nn", "ss", "vv", "33"}, "change"))
	}
}

func TestParse(t *testing.T) {
	ini := asserts.Must1(parser.Parse("",
		strings.NewReader(url1),
		participle.Trace(os.Stdout),
	))

	pretty.Println(ini)
	pp := ParseToRoute(ini)
	asserts.If(!reflect.DeepEqual(pp.Paths, []string{"v1", "users", "hh", "*", "hello", "*", "*", "hhh", "*", "messages", "nn", "ss", "**"}), "not match")
	asserts.If(*pp.Verb != "change", "not match")
	pretty.Println(pp)

	t.Log(pp.String())
	p1 := asserts.Must1(parser.Parse("", strings.NewReader(pp.String())))
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

func TestRouteTree(t *testing.T) {
	var tree = NewRouteTree()
	asserts.Exit(tree.Add("get", url1, "get_test"))
	pp, err := tree.Match("get", "/v1/users/hh/1111/hello/444/555/hhh/6666/messages/nn/ss/mm/ddd/44:change")
	asserts.Exit(err)
	pretty.Println(pp)
	pretty.Println(tree.nodes)
}
