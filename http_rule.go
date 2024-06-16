package http_rule

import (
	"fmt"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/pubgo/funk/generic"
)

const (
	DoubleStar = "**"
	Star       = "*"
)

var (
	lex = assert.Exit1(lexer.NewSimple([]lexer.SimpleRule{
		{Name: "Ident", Pattern: `[a-zA-Z]\w*`},
		{Name: "Punct", Pattern: `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
	}))

	parser = assert.Exit1(participle.Build[httpRule](
		participle.Lexer(lex),
	))
)

//     http rule
//     Template = "/" Segments [ Verb ] ;
//     Segments = Segment { "/" Segment } ;
//     Segment  = "*" | "**" | LITERAL | Variable ;
//     Variable = "{" FieldPath [ "=" Segments ] "}" ;
//     FieldPath = IDENT { "." IDENT } ;
//     Verb     = ":" LITERAL ;

type httpRule struct {
	Slash    string    `@"/"`
	Segments *segments `@@!`
	Verb     *string   `(":" @Ident)?`
}

type segments struct {
	Segments []*segment `@@ ("/" @@)*`
}

type segment struct {
	Path     *string   `@("*" "*" | "*" | Ident)`
	Variable *variable `| @@`
}

type variable struct {
	Fields   []string  `"{" @Ident ("." @Ident)*`
	Segments *segments `("=" @@)? "}"`
}

type pathVariable struct {
	Fields     []string
	start, end int
}

type RoutePath struct {
	Paths []string
	Verb  *string
	Vars  []*pathVariable
}

type PathVar struct {
	Fields []string
	Value  string
}

func (r RoutePath) Match(urls []string, verb string) ([]PathVar, error) {
	if len(urls) < len(r.Paths) {
		return nil, errors.New("urls length not match")
	}

	if r.Verb != nil {
		if generic.FromPtr(r.Verb) != verb {
			return nil, errors.New("verb not match")
		}
	}

	for i := range r.Paths {
		path := r.Paths[i]
		if path == Star {
			continue
		}

		if path == urls[i] {
			continue
		}

		if path == DoubleStar {
			continue
		}

		return nil, errors.New("path is not match")
	}

	var vv []PathVar
	for _, v := range r.Vars {
		pathVar := PathVar{Fields: v.Fields}
		if v.end > 0 {
			pathVar.Value = strings.Join(urls[v.start:v.end+1], "/")
		} else {
			pathVar.Value = strings.Join(urls[v.start:], "/")
		}

		vv = append(vv, pathVar)
	}

	return vv, nil
}

func (r RoutePath) String() string {
	url := "/"

	paths := make([]string, len(r.Paths))
	copy(paths, r.Paths)

	for _, v := range r.Vars {
		varS := "{" + strings.Join(v.Fields, ".") + "="
		end := generic.Ternary(v.end == -1, len(paths)-1, v.end)

		for i := v.start; i <= end; i++ {
			varS += generic.Ternary(i == v.start, paths[i], "/"+paths[i])
			if i > v.start {
				paths[i] = ""
			}
		}

		varS += "}"
		paths[v.start] = varS
	}

	url += strings.Join(generic.Filter(paths, func(s string) bool { return s != "" }), "/")

	if r.Verb != nil {
		url += ":" + generic.FromPtr(r.Verb)
	}

	return url
}

func handleSegments(s *segment, rr *RoutePath) {
	if s.Path != nil {
		rr.Paths = append(rr.Paths, *s.Path)
		return
	}

	vv := &pathVariable{Fields: s.Variable.Fields, start: len(rr.Paths)}
	if s.Variable.Segments == nil {
		rr.Paths = append(rr.Paths, Star)
	} else {
		for _, v := range s.Variable.Segments.Segments {
			handleSegments(v, rr)
		}
	}

	if len(rr.Paths) > 0 && rr.Paths[len(rr.Paths)-1] == DoubleStar {
		vv.end = -1
	} else {
		vv.end = len(rr.Paths) - 1
	}

	rr.Vars = append(rr.Vars, vv)
}

func ParseToRoute(rule *httpRule) *RoutePath {
	r := new(RoutePath)
	r.Verb = rule.Verb

	if rule.Segments != nil {
		for _, v := range rule.Segments.Segments {
			handleSegments(v, r)
		}
	}

	return r
}

func Parse(url string) (*httpRule, error) {
	return parser.ParseString("", url)
}

func NewRouteTree() *RouteTree {
	return &RouteTree{nodes: make(map[string]*PathNode)}
}

type RouteTree struct {
	nodes map[string]*PathNode
}

func (r *RouteTree) Add(method string, url string, operation string) error {
	rule, err := Parse(url)
	if err != nil {
		return err
	}

	var node = ParseToRoute(rule)
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

	var getVars = func(vars []*pathVariable, urls []string) []PathVar {
		var vv = make([]PathVar, 0, len(vars))
		for _, v := range vars {
			pathVar := PathVar{Fields: v.Fields}
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
	Vars      []PathVar
}
