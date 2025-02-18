package main

import (
	_ "embed"
	"encoding/xml"
	"fmt"
	"html/template"
	"os"

	"github.com/Masterminds/sprig/v3"
	"github.com/alecthomas/kong"
	"github.com/davecgh/go-spew/spew"
	"github.com/jstemmer/go-junit-report/formatter"
)

type CSSPrinter struct {
}

func (c *CSSPrinter) Run(cxt *kong.Context) error {
	fmt.Println(styles)
	return nil
}

type HTML struct {
}

func (h *HTML) Run(cxt *kong.Context) error {
	fmt.Println(report)
	return nil
}

type Generator struct {
	JunitXML *os.File `arg:"" help:"Path to the JUnit XML file to generate a report from. use '-' for stdin."`
}

func (g *Generator) Run(ctx *kong.Context) error {
	defer g.JunitXML.Close()
	ctx.FatalIfErrorf(xml.NewDecoder(g.JunitXML).Decode(&suites))
	tmplCtx := struct {
		Suites      []formatter.JUnitTestSuite
		NumFailures int
		NumTotal    int
		CSS         template.CSS
	}{
		Suites: suites.Suites,
		CSS:    template.CSS(styles),
	}
	if cfg.Debug {
		spew.Dump(tmplCtx)
		spew.Dump(suites)
	}

	for _, s := range suites.Suites {
		tmplCtx.NumFailures += s.Failures
		tmplCtx.NumTotal += len(s.TestCases)
	}
	return tmpl.Execute(os.Stdout, tmplCtx)
}

var (
	//go:embed style.css
	styles string
	//go:embed report.gohtml
	report string
	tmpl   = template.Must(template.New("report").Funcs(sprig.FuncMap()).Parse(report))
	suites formatter.JUnitTestSuites
	cfg    struct {
		Debug bool `short:"v" long:"debug" description:"Show debug information"`
		Print struct {
			// Make this a dumper with args
			CSS  CSSPrinter `cmd:"css" help:"Print the embedded CSS"`
			HTML HTML       `cmd:"html" help:"Print the embedded HTML"`
		} `cmd:"print" help:"Print the embedded resources"`
		Generate Generator `cmd:"generate" help:"Generate a report from the input XML"`
	}
)

func main() {

	ctx := kong.Parse(&cfg)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
