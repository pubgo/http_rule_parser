package http_rule

import (
	"fmt"
	"strings"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
)

func NewRouteTree() *RouteTree {
	return &RouteTree{nodes: make(map[string]*PathNode)}
}

type RouteTree struct {
	nodes map[string]*PathNode
}

func (r *RouteTree) Add(method string, url string, operation string) error {
	rule, err := parse(url)
	if err != nil {
		return err
	}

	var node = parseToRoute(rule)
	if len(node.Paths) == 0 {
		return fmt.Errorf("path is null")
	}

	var nodes = r.nodes
	for i, n := range node.Paths {
		var lastNode = nodes[n]
		if lastNode == nil {
			lastNode = &PathNode{
				nodes: make(map[string]*PathNode),
				verbs: make(map[string]*RouteTarget),
			}
			nodes[n] = lastNode
		}
		nodes = lastNode.nodes

		if i == len(node.Paths)-1 {
			lastNode.verbs[generic.FromPtr(node.Verb)] = &RouteTarget{
				Method:    method,
				Operation: &operation,
				Verb:      &method,
				Vars:      node.Vars,
			}
		}
	}
	return nil
}

func (r *RouteTree) Match(method, url string) (*MatchPath, error) {
	var urls = strings.Split(strings.Trim(strings.TrimSpace(url), "/"), "/")
	var lastPath = strings.SplitN(urls[len(urls)-1], ":", 2)
	var verb = ""

	urls[len(urls)-1] = lastPath[0]
	if len(lastPath) > 1 {
		verb = lastPath[1]
	}

	var getVars = func(vars []*pathVariable, urls []string) []pathVar {
		var vv = make([]pathVar, 0, len(vars))
		for _, v := range vars {
			pathVar := pathVar{Fields: v.Fields}
			if v.end > 0 {
				pathVar.Value = strings.Join(urls[v.start:v.end+1], "/")
			} else {
				pathVar.Value = strings.Join(urls[v.start:], "/")
			}

			vv = append(vv, pathVar)
		}
		return vv
	}
	var getPath = func(nodes map[string]*PathNode, names ...string) *PathNode {
		for _, n := range names {
			path := nodes[n]
			if path != nil {
				return path
			}
		}
		return nil
	}

	var nodes = r.nodes
	for _, n := range urls {
		path := getPath(nodes, n, Star, DoubleStar)
		if path == nil {
			return nil, errors.Format("path node not match, node=%s", n)
		}

		if vv := path.verbs[verb]; vv != nil && vv.Operation != nil && vv.Method == method {
			return &MatchPath{
				Operation: generic.FromPtr(vv.Operation),
				Verb:      verb,
				Vars:      getVars(vv.Vars, urls),
			}, nil
		}
		nodes = path.nodes
	}

	return nil, errors.New("path not match")
}

type RouteTarget struct {
	Method    string
	Operation *string
	Verb      *string
	Vars      []*pathVariable
}

type PathNode struct {
	nodes map[string]*PathNode
	verbs map[string]*RouteTarget
}

type MatchPath struct {
	Operation string
	Verb      string
	Vars      []pathVar
}
