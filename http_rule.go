package http_rule

import (
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/pubgo/funk/generic"
)

const (
	StarStar = "**"
	Star     = "*"
)

var (
	lex = assert.Exit1(lexer.NewSimple([]lexer.SimpleRule{
		{Name: "Ident", Pattern: `[a-zA-Z]\w*`},
		{Name: "Punct", Pattern: `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
	}))

	parser = assert.Exit1(participle.Build[HttpRule](
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

type HttpRule struct {
	Slash    string    `@"/"`
	Segments *Segments `@@!`
	Verb     *string   `(":" @Ident)?`
}

type Segments struct {
	Segments []*Segment `@@ ("/" @@)*`
}

type Segment struct {
	Path     *string   `@("*" "*" | "*" | Ident)`
	Variable *Variable `| @@`
}

type Variable struct {
	Fields   []string  `"{" @Ident ("." @Ident)*`
	Segments *Segments `("=" @@)? "}"`
}

type pathVariable struct {
	Fields     []string
	start, end int
}

type Route struct {
	Paths []string
	Verb  *string
	Vars  []*pathVariable
}

type PathVar struct {
	Fields []string
	Value  string
}

func (r Route) Match(urls []string, verb string) ([]PathVar, error) {
	if len(urls) < len(r.Paths) {
		return nil, errors.New("urls length is less than path length")
	}

	if r.Verb != nil {
		if generic.FromPtr(r.Verb) != verb {
			return nil, errors.New("verb is not match")
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

		if path == StarStar {
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

func (r Route) String() string {
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

func handleSegments(s *Segment, rr *Route) {
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

	if len(rr.Paths) > 0 && rr.Paths[len(rr.Paths)-1] == StarStar {
		vv.end = -1
	} else {
		vv.end = len(rr.Paths) - 1
	}

	rr.Vars = append(rr.Vars, vv)
}

func ParseToRoute(rule *HttpRule) *Route {
	r := new(Route)
	r.Verb = rule.Verb

	if rule.Segments != nil {
		for _, v := range rule.Segments.Segments {
			handleSegments(v, r)
		}
	}

	return r
}

func Parse(url string) (*HttpRule, error) {
	return parser.ParseString("", url)
}
