# protobuf http rule parser

## example 

```go
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
```

```go
&http_rule.HttpRule{
  Slash:    "/",
  Segments: &http_rule.Segments{
    Segments: []*http_rule.Segment{
      &http_rule.Segment{
        Path:     &"v1",
        Variable: (*http_rule.Variable)(nil),
      },
      &http_rule.Segment{
        Path:     &"users",
        Variable: (*http_rule.Variable)(nil),
      },
      &http_rule.Segment{
        Path:     (*string)(nil),
        Variable: &http_rule.Variable{
          Fields: []string{
            "aa",
            "ba",
            "ca",
          },
          Segments: &http_rule.Segments{
            Segments: []*http_rule.Segment{
              &http_rule.Segment{
                Path:     &"hh",
                Variable: (*http_rule.Variable)(nil),
              },
              &http_rule.Segment{
                Path:     &"*",
                Variable: (*http_rule.Variable)(nil),
              },
            },
          },
        },
      },
      &http_rule.Segment{
        Path:     &"hello",
        Variable: (*http_rule.Variable)(nil),
      },
      &http_rule.Segment{
        Path:     (*string)(nil),
        Variable: &http_rule.Variable{
          Fields: []string{
            "hello",
            "abc",
          },
          Segments: (*http_rule.Segments)(nil),
        },
      },
      &http_rule.Segment{
        Path:     &"messages",
        Variable: (*http_rule.Variable)(nil),
      },
      &http_rule.Segment{
        Path:     (*string)(nil),
        Variable: &http_rule.Variable{
          Fields: []string{
            "messageId",
          },
          Segments: &http_rule.Segments{
            Segments: []*http_rule.Segment{
              &http_rule.Segment{
                Path:     &"nn",
                Variable: (*http_rule.Variable)(nil),
              },
              &http_rule.Segment{
                Path:     &"ss",
                Variable: (*http_rule.Variable)(nil),
              },
              &http_rule.Segment{
                Path:     &"**",
                Variable: (*http_rule.Variable)(nil),
              },
            },
          },
        },
      },
    },
  },
  Verb: &"change",
}


&http_rule.Route{
  Paths: []string{
    "v1",
    "users",
    "hh",
    "*",
    "hello",
    "*",
    "messages",
    "nn",
    "ss",
    "**",
  },
  Verb: &"change",
  Vars: []*http_rule.pathVariable{
    &http_rule.pathVariable{
      Fields: []string{
        "aa",
        "ba",
        "ca",
      },
      start: 2,
      end:   3,
    },
    &http_rule.pathVariable{
      Fields: []string{
        "hello",
        "abc",
      },
      start: 5,
      end:   5,
    },
    &http_rule.pathVariable{
      Fields: []string{
        "messageId",
      },
      start: 7,
      end:   -1,
    },
  },
}
```