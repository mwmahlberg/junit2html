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
	"github.com/pkg/errors"
)

type printer struct {
	Type string `arg:"" help:"Type of resource to print." name:"css|html" enum:"css,html"`
}

func (p *printer) Run(ctx context) error {
	switch p.Type {
	case "css":
		fmt.Println(defaultStyles)
	case "html":
		fmt.Println(defaultTemplate)
	}
	return nil
}

type generator struct {
	JunitXML *os.File `arg:"" help:"Path to the JUnit XML file to generate a report from. Use '-' for stdin."`
	Template string   `short:"t" type:"existingfile" help:"Use an alternate template." placeholder:"/path/to/go.tmpl" group:"Rendering options"`
	CSS      string   `short:"s" name:"styles" type:"filecontent" help:"Use an alternate CSS file." placeholder:"/path/to.css" group:"Rendering options"`
}

func (g *generator) Run(ctx *context) (err error) {
	defer g.JunitXML.Close()
	tmpl := template.Must(template.New("report").Funcs(sprig.FuncMap()).Parse(ctx.defaultTemplate))

	if g.Template != "" {
		tmpl, err = template.New("report").Funcs(sprig.FuncMap()).ParseFiles(g.Template)
		if err != nil {
			return errors.Wrap(err, "failed to parse template")
		}
	}

	var suites formatter.JUnitTestSuites
	if err = xml.NewDecoder(g.JunitXML).Decode(&suites); err != nil {
		return errors.Wrap(err, "failed to decode JUnit XML")
	}

	tmplCtx := struct {
		Suites []formatter.JUnitTestSuite
		CSS    template.CSS
	}{
		Suites: suites.Suites,
		CSS:    template.CSS(defaultStyles),
	}

	if cfg.Debug {
		spew.Dump(tmplCtx)
		spew.Dump(suites)
	}
	return tmpl.ExecuteTemplate(os.Stdout, "report", tmplCtx)
}

var (
	//go:embed style.css
	defaultStyles string
	//go:embed report.gohtml
	defaultTemplate string
	// suites          formatter.JUnitTestSuites
	cfg struct {
		Debug    bool      `short:"d" long:"debug" help:"Show debug information"`
		Print    printer   `cmd:"print" help:"Print an embedded resource"`
		Generate generator `cmd:"generate" help:"Generate a report from the input XML"`
	}
)

type context struct {
	defaultTemplate string
	defaultStyles   string
}

func main() {

	ctx := kong.Parse(&cfg,
		kong.Name("junit-report"),
		kong.Description("Generate a HTML report from JUnit XML"))
	err := ctx.Run(&context{defaultStyles: defaultStyles, defaultTemplate: defaultTemplate})
	ctx.FatalIfErrorf(err)
}
