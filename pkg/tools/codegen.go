// +build ignore

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/jnpr-tjiang/echo-apisvr/pkg/utils"
	"gopkg.in/yaml.v2"
)

type relationshipInfo struct {
	Entity       string `yaml:"entity"`
	RelationType string `yaml:"type"`

	// derived fields
	FieldName    string `yaml:"-"`
	MappingTable string `yaml:"-"`
}

type normalizedField struct {
	Field     string `yaml:"field"`
	FieldType string `yaml:"type"`
	Indexed   bool   `yaml:"indexed"`

	// derived fields
	ColumnName string `yaml:"-"`
}

type entityInfo struct {
	Entity           string             `yaml:"entity"`
	Parents          []string           `yaml:"parents"`
	Relationships    []relationshipInfo `yaml:"relationships"`
	NormalizedFields []normalizedField  `yaml:"normalize"`
	MappingEntity    bool               `yaml:"-"`
	FromEntity       string             `yaml:"-"`
	ToEntity         string             `yaml:"-"`
}

func addMoreInfo(entities []entityInfo) []entityInfo {
	mappingEntities := []entityInfo{}
	for i := range entities {
		entity := &entities[i]
		for j := range entity.Relationships {
			relation := &entity.Relationships[j]
			relation.FieldName = utils.Pluralize(relation.Entity)
			if relation.RelationType == "many2many" {
				relation.MappingTable = strings.ToLower(entity.Entity + "_" + relation.FieldName)
				mappingEntities = append(mappingEntities, entityInfo{
					Entity:        entity.Entity + relation.Entity,
					Parents:       []string{},
					FromEntity:    entity.Entity,
					ToEntity:      relation.Entity,
					MappingEntity: true,
				})
			}
		}
		for j := range entity.NormalizedFields {
			field := &entity.NormalizedFields[j]
			field.ColumnName = strings.ToLower(field.Field)
		}
	}
	return append(entities, mappingEntities...)
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// fmt.Printf("Running %s go on %s\n", os.Args[0], os.Getenv("GOFILE"))
	// fmt.Printf("  os.Args = %#v\n", os.Args)

	cwd, err := os.Getwd()
	die(err)

	modelFilePath := cwd + "/" + os.Args[1]
	yamlFile, err := ioutil.ReadFile(modelFilePath)
	die(err)

	entities := []entityInfo{}
	err = yaml.Unmarshal(yamlFile, &entities)
	die(err)

	entities = addMoreInfo(entities)

	tmplFilePath := path.Join(cwd, "model.tmpl")
	// tmplFilePath := path.Join(cwd, "../models/model.tmpl")
	_, err = os.Stat(tmplFilePath)
	die(err)

	outFilePath := path.Join(cwd, os.Args[2])
	f, err := os.Create(outFilePath)
	die(err)
	defer f.Close()

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}
	tmplName := path.Base(tmplFilePath)
	tpl := template.Must(
		template.New(tmplName).Funcs(funcMap).ParseFiles(tmplFilePath))
	err = tpl.Execute(f, entities)
	if err != nil {
		log.Printf("Failed to resolve template: %s", err)
	}

	// for _, ev := range []string{"GOARCH", "GOOS", "GOFILE", "GOLINE", "GOPACKAGE", "DOLLAR"} {
	// 	fmt.Println("  ", ev, "=", os.Getenv(ev))
	// }
}
