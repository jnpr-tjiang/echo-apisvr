package models

import (
	"testing"

	"github.com/jnpr-tjiang/echo-apisvr/pkg/utils"
)

func TestBaseModel_New(t *testing.T) {
	Init()
	entity, err := NewEntity("domain")
	if err != nil {
		t.Error(err)
		return
	}
	entityType := utils.TypeOf(entity)
	if entityType != "Domain" {
		t.Errorf("Expected value: Domain but got %s", entityType)
	}
}

func TestBaseModel_ModelNames(t *testing.T) {
	Init()
	names := ModelNames()
	want := []string{"domain", "project", "device", "devicefamily"}
	if len(names) != len(want) {
		t.Errorf("Expect %v but got %v", want, names)
	}
	for _, v := range names {
		if _, ok := utils.Find(want, v); !ok {
			t.Errorf("Expect %v but got %v", want, names)
		}
	}
}
