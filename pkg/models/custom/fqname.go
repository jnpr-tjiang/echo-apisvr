package custom

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

// FQName is the fully qualified unique name
type FQName []string

// Value return json value, implement driver.Valuer interface
func (fqn FQName) Value() (driver.Value, error) {
	if len(fqn) == 0 {
		return nil, nil
	}
	str := "["
	for i, v := range fqn {
		if (i + 1) != len(fqn) {
			str += fmt.Sprintf("\"%s\", ", v)
		} else {
			str += fmt.Sprintf("\"%s\"]", v)
		}
	}
	return str, nil
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (fqn *FQName) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal FRName value:", value))
	}
	str = strings.Trim(str, "[]")
	str = strings.ReplaceAll(str, "\"", "")
	*fqn = strings.Split(str, ", ")
	return nil
}

// Parent returns the parent FQName
func (fqn FQName) Parent() FQName {
	if len(fqn) <= 1 {
		return FQName{}
	} else {
		return fqn[:(len(fqn) - 1)]
	}
}

// ParseParentFQName extract out the parent FQName from a FQName
func ParseParentFQName(fqnStr string) (string, error) {
	fqn := FQName{}
	if err := fqn.Scan(fqnStr); err != nil {
		return "", err
	}
	parentFQN, err := fqn.Parent().Value()
	if parentFQN == nil {
		return "", err
	} else {
		return parentFQN.(string), err
	}
}

// ConstructFQName from parent FQName
func ConstructFQName(parentFQName string, name string) string {
	if parentFQName == "" {
		return fmt.Sprintf("[\"%s\"]", name)
	}
	return fmt.Sprintf(`%s, "%s"]`, parentFQName[:len(parentFQName)-1], name)
}
