package http_rule_parser

import (
	asserts "github.com/pubgo/funk/assert"
)

func ExampleRouteTree() {
	var tree = NewRouteTree()
	asserts.Exit(tree.Add("get", url1, "get_test"))
	pp, err := tree.Match("get", "/v1/users/hh/1111/hello/444/555/hhh/6666/messages/nn/ss/mm/ddd/44:change")
	asserts.Exit(err)
	asserts.If(pp.Operation != "get_test", "not match")
	// Output:
}
