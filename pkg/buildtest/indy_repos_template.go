package buildtest

import (
	"bytes"
	"log"
	"text/template"
)

var isNotLast = template.FuncMap{
	// The name "inc" is what the function will be called in the template text.
	"isNotLast": func(index int, array []string) bool {
		return index < len(array)-1
	},
}

// IndyGroupVars ...
type IndyGroupVars struct {
	Name         string
	Type         string
	Constituents []string
}

// IndyGroupTemplate ...
func IndyGroupTemplate(indyGroupVars *IndyGroupVars) string {
	groupTemplate := `{
  "type" : "group",
  "key" : "{{.Type}}:group:{{.Name}}",
  "metadata" : {
    "changelog" : "init group {{.Name}}"
  },
  "disabled" : false,
  "constituents" : [{{range $index,$con := .Constituents}}"{{$con}}"{{if isNotLast $index $.Constituents}},{{end}}{{end}}],
  "packageType" : "{{.Type}}",
  "name" : "{{.Name}}",
  "type" : "group",
  "disable_timeout" : 0,
  "path_style" : "plain",
  "authoritative_index" : false,
  "prepend_constituent" : false
}`

	t := template.Must(template.New("settings").Funcs(isNotLast).Parse(groupTemplate))
	var buf bytes.Buffer
	err := t.Execute(&buf, indyGroupVars)
	if err != nil {
		log.Fatal("executing template:", err)
	}

	return buf.String()
}

// IndyHostedVars ...
type IndyHostedVars struct {
	Name string
	Type string
}

// IndyHostedTemplate ...
func IndyHostedTemplate(indyHostedVars *IndyHostedVars) string {
	hostedTemplate := `{
  "key" : "{{.Type}}:hosted:{{.Name}}",
  "description" : "{{.Name}}",
  "metadata" : {
    "changelog" : "init hosted {{.Name}}"
  },
  "disabled" : false,
  "snapshotTimeoutSeconds" : 0,
  "readonly" : false,
  "packageType" : "{{.Type}}",
  "name" : "{{.Name}}",
  "type" : "hosted",
  "disable_timeout" : 0,
  "path_style" : "plain",
  "authoritative_index" : true,
  "allow_snapshots" : true,
  "allow_releases" : true
}`

	t := template.Must(template.New("settings").Parse(hostedTemplate))
	var buf bytes.Buffer
	err := t.Execute(&buf, indyHostedVars)
	if err != nil {
		log.Fatal("executing template:", err)
	}

	return buf.String()
}
