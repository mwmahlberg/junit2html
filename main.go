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

type Printer struct {
	Type string `arg:"" help:"Type of resource to print." name:"css|html" enum:"css,html"`
}

func (p *Printer) Run(ctx *kong.Context) error {
	switch p.Type {
	case "css":
		fmt.Println(styles)
	case "html":
		fmt.Println(templateData)
	}
	return nil
}

type Generator struct {
	JunitXML *os.File `arg:"" help:"Path to the JUnit XML file to generate a report from. use '-' for stdin."`
}

func (g *Generator) Run(ctx *kong.Context) error {
	defer g.JunitXML.Close()
	ctx.FatalIfErrorf(xml.NewDecoder(g.JunitXML).Decode(&suites))
	tmplCtx := struct {
		Suites []formatter.JUnitTestSuite
		CSS    template.CSS
	}{
		Suites: suites.Suites,
		CSS:    template.CSS(styles),
	}
	if cfg.Debug {
		spew.Dump(tmplCtx)
		spew.Dump(suites)
	}

	return tmpl.Execute(os.Stdout, tmplCtx)
}

var (
	//go:embed style.css
	styles string
	//go:embed report.gohtml
	templateData string
	tmpl         = template.Must(template.New("report").Funcs(sprig.FuncMap()).Parse(templateData))
	suites       formatter.JUnitTestSuites
	cfg          struct {
		Debug    bool      `short:"v" long:"debug" description:"Show debug information"`
		Print    Printer   `cmd:"print" help:"Print an embedded resource"`
		Generate Generator `cmd:"generate" help:"Generate a report from the input XML"`
	}
)

func main() {

	ctx := kong.Parse(&cfg)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
