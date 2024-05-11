package http_rule

import (
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/pretty"
)

func TestName(t *testing.T) {
	ini := assert.Must1(parser.Parse("",
		strings.NewReader("/v1/users/{aa.ba.ca=hh/*}/hello/{hello.abc}/messages/{messageId=nn/ss/**}:change"),
		participle.Trace(os.Stdout),
	))

	pretty.Println(ini)
	pp := ParseToRoute(ini)
	pretty.Println(pp)

	p1 := assert.Must1(parser.Parse("", strings.NewReader(pp.String())))
	if pp2 := ParseToRoute(p1).String(); pp2 != pp.String() {
		t.Fatal("not equal", pp2, pp.String())
	}
}
