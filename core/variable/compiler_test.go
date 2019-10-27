package variable_test

import (
	"github.com/axetroy/s4/core/variable"
	"reflect"
	"testing"
)

func TestCompile(t *testing.T) {
	type args struct {
		template string
		varMap   map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{
				template: "hello {{name}}",
				varMap: map[string]string{
					"name": "test",
				},
			},
			want: "hello test",
		},
		{
			name: "multiple variables",
			args: args{
				template: "hello {{name}}, I am {{age}} years old",
				varMap: map[string]string{
					"name": "test",
					"age":  "18",
				},
			},
			want: "hello test, I am 18 years old",
		},
		{
			name: "missing variables",
			args: args{
				template: "hello {{name}}, I am {{age}} years old. I live in {{city}}",
				varMap: map[string]string{
					"name": "test",
					"age":  "18",
					//"city": "CK", // the missing variable
				},
			},
			want: "hello test, I am 18 years old. I live in ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := variable.Compile(tt.args.template, tt.args.varMap); got != tt.want {
				t.Errorf("Compile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompileArray(t *testing.T) {
	type args struct {
		templates []string
		varMap    map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "basic",
			args: args{
				templates: []string{"hello {{name}}"},
				varMap: map[string]string{
					"name": "test",
				},
			},
			want: []string{"hello test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := variable.CompileArray(tt.args.templates, tt.args.varMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CompileArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
