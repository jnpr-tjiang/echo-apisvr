package custom

import (
	"database/sql/driver"
	"reflect"
	"testing"
)

func TestFQName_Value(t *testing.T) {
	tests := []struct {
		name    string
		fqn     FQName
		want    driver.Value
		wantErr bool
	}{
		{
			name:    "SingleNameFQName",
			fqn:     FQName{"default"},
			want:    "[\"default\"]",
			wantErr: false,
		},
		{
			name:    "MultiNameFQName",
			fqn:     FQName{"default", "juniper", "junos"},
			want:    "[\"default\", \"juniper\", \"junos\"]",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fqn.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("FQName.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FQName.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFQName_Scan(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		fqn     *FQName
		args    args
		want    FQName
		wantErr bool
	}{
		{
			name: "SingleNameFQN",
			fqn:  &FQName{},
			args: args{
				value: "[\"default\"]",
			},
			want: FQName{"default"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fqn.Scan(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("FQName.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(*tt.fqn) != len(tt.want) {
				t.Errorf("expect value is %v but got %v", tt.want, *tt.fqn)
			}
			for i := 0; i < len(tt.want); i++ {
				if (*tt.fqn)[i] != tt.want[i] {
					t.Errorf("expect value is %v but got %v", tt.want, *tt.fqn)
				}
			}
		})
	}
}

func TestFQName_Parent(t *testing.T) {
	tests := []struct {
		name string
		fqn  FQName
		want FQName
	}{
		{
			name: "SingleNameFQN",
			fqn:  FQName{"default"},
			want: FQName{},
		},
		{
			name: "MultiNameFQN",
			fqn:  FQName{"default", "juniper", "junos"},
			want: FQName{"default", "juniper"},
		},
		{
			name: "EmptyFQN",
			fqn:  FQName{},
			want: FQName{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fqn.Parent(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FQName.Parent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseParentFQName(t *testing.T) {
	type args struct {
		fqnStr string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 error
	}{
		{
			name: "SingleNameFQN",
			args: args{
				fqnStr: "[\"default\"]",
			},
			want:  "",
			want1: nil,
		},
		{
			name: "MultiNameFQN",
			args: args{
				fqnStr: "[\"default\", \"juniper\", \"junos\"]",
			},
			want:  "[\"default\", \"juniper\"]",
			want1: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ParseParentFQName(tt.args.fqnStr)
			if got != tt.want {
				t.Errorf("ParseParentFQName() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ParseParentFQName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestConstructFQName(t *testing.T) {
	type args struct {
		parentFQName string
		name         string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "EmptyParentFQN",
			args: args{
				parentFQName: "",
				name:         "default",
			},
			want: "[\"default\"]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConstructFQName(tt.args.parentFQName, tt.args.name); got != tt.want {
				t.Errorf("ConstructFQName() = %v, want %v", got, tt.want)
			}
		})
	}
}
