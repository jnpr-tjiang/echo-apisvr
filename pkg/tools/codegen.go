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
	ExtendBase       bool               `yaml:"extendBase"`
	Abstract         bool               `yaml:"abstract"`
	Parents          []string           `yaml:"parents"`
	Relationships    []relationshipInfo `yaml:"relationships"`
	NormalizedFields []normalizedField  `yaml:"normalize"`
}

func addMoreInfo(entities []entityInfo) []entityInfo {
	// mappingEntities := []entityInfo{}
	for i := range entities {
		entity := &entities[i]
		for j := range entity.Relationships {
			relation := &entity.Relationships[j]
			relation.FieldName = utils.Pluralize(relation.Entity)
			if relation.RelationType == "many2many" {
				relation.MappingTable = strings.ToLower(entity.Entity + "_" + relation.FieldName)
			}
		}
		for j := range entity.NormalizedFields {
			field := &entity.NormalizedFields[j]
			field.ColumnName = strings.ToLower(field.Field)
		}
	}
	return entities
}

func main() {
	// fmt.Printf("Running %s go on %s\n", os.Args[0], os.Getenv("GOFILE"))
	// fmt.Printf("  os.Args = %#v\n", os.Args)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	modelFilePath := cwd + "/" + os.Args[1]
	yamlFile, err := ioutil.ReadFile(modelFilePath)
	if err != nil {
		log.Printf("Failed to open file: %s", modelFilePath)
		panic(err)
	}

	entities := []entityInfo{}
	err = yaml.Unmarshal(yamlFile, &entities)
	if err != nil {
		log.Printf("Invalid yaml file: %s", os.Args[1])
		panic(err)
	}
	addMoreInfo(entities)

	// tmplFilePath := path.Join(cwd, "model.tmpl")
	tmplFilePath := path.Join(cwd, "../models/model.tmpl")
	_, err = os.Stat(tmplFilePath)
	if err != nil {
		log.Printf("Template file not found: %s", err)
		panic(err)
	}

	tpl := template.Must(template.New(path.Base(tmplFilePath)).ParseFiles(tmplFilePath))
	err = tpl.Execute(os.Stdout, entities)
	if err != nil {
		log.Printf("Failed to resolve template: %s", err)
	}

	// for _, ev := range []string{"GOARCH", "GOOS", "GOFILE", "GOLINE", "GOPACKAGE", "DOLLAR"} {
	// 	fmt.Println("  ", ev, "=", os.Getenv(ev))
	// }
}
