package main

import (
	"bytes"
	"io"
	"os"
	"sort"
	"text/template"

	"github.com/mitchellh/cli"
)

// CommandDoc stores command documentation strings.
type CommandDoc struct {
	Name     string
	Synopsis string
	Help     string
}

func runGodoc(commands map[string]cli.CommandFactory) int {
	f, err := os.Create("doc.go")
	if err != nil {
		return 1
	}
	defer f.Close()

	commandDocs := newCommandDocs(commands)

	var buf bytes.Buffer
	if err := tmpl(&buf, docTmpl, commandDocs); err != nil {
		return 1
	}

	err = tmpl(f, godocTmpl, struct {
		Content string
	}{
		Content: buf.String(),
	})
	if err != nil {
		return 1
	}
	return 0
}

func newCommandDocs(commands map[string]cli.CommandFactory) []CommandDoc {
	keys := make([]string, 0, len(commands))

	// Extract key (command name) for sort.
	for key, _ := range commands {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	commandDocs := make([]CommandDoc, 0, len(commands))
	for _, key := range keys {
		if key == "version" {
			continue
		}
		cmdFunc, ok := commands[key]
		if !ok {
			// Should not reach here..
			panic("command not found: " + key)
		}

		cmd, _ := cmdFunc()
		commandDocs = append(commandDocs, CommandDoc{
			Name:     key,
			Synopsis: cmd.Synopsis(),
			Help:     cmd.Help(),
		})
	}

	return commandDocs
}

// tmpl evaluates template content with data
// and write them to writer, return error if any
func tmpl(wr io.Writer, content string, data interface{}) error {

	tmpl, err := template.New("doc").Parse(content)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(wr, data); err != nil {
		return err
	}

	return nil
}

var docTmpl = `Command gcli generates a skeleton (codes and its directory structure) you need to start building CLI tool by Golang.
https://github.com/tcnksm/gcli

Usage:

    gcli [-version] [-help]  <command> [<options>]

Available commands:
{{ range .}}
    {{ .Name | printf "%-11s"}} {{ .Synopsis }}{{end}}

Use "gcli <command> -help" for more information about command.

{{ range . }}

{{ .Synopsis }}

{{ .Help }}

{{ end }}
`

var godocTmpl = `// DO NOT EDIT THIS FILE.
// THIS FILE IS GENERATED BY GO GENERATE.

/*
{{ .Content }}
*/
package main
`
