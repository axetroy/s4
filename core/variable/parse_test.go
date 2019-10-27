package variable_test

import (
	"github.com/axetroy/s4/core/variable"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    *variable.Variable
		wantErr bool
	}{
		{
			name: "basic literal",
			args: args{
				input: "PRIVATE_KEY = 123",
			},
			want: &variable.Variable{
				Key:    "PRIVATE_KEY",
				Value:  "123",
				Type:   variable.TypeLiteral,
				Remote: false,
			},
		},
		{
			name: "basic env",
			args: args{
				input: "PRIVATE_KEY = $GOPATH:local",
			},
			want: &variable.Variable{
				Key:    "PRIVATE_KEY",
				Value:  "GOPATH",
				Type:   variable.TypeEnv,
				Remote: false,
			},
		},
		{
			name: "basic env with remote",
			args: args{
				input: "PRIVATE_KEY = $GOPATH:remote",
			},
			want: &variable.Variable{
				Key:    "PRIVATE_KEY",
				Value:  "GOPATH",
				Type:   variable.TypeEnv,
				Remote: true,
			},
		},
		{
			name: "basic local command",
			args: args{
				input: `VERSION <= ["npm", "version"]`,
			},
			want: &variable.Variable{
				Key:    "VERSION",
				Value:  `npm version`,
				Type:   variable.TypeCommand,
				Remote: false,
			},
		},
		{
			name: "basic remote command",
			args: args{
				input: `VERSION <= npm version`,
			},
			want: &variable.Variable{
				Key:    "VERSION",
				Value:  `npm version`,
				Type:   variable.TypeCommand,
				Remote: true,
			},
		},
		{
			name: "Invalid syntax",
			args: args{
				input: `VERSION *= npm version`,
			},
			wantErr: true,
		},
		{
			name: "invalid tag",
			args: args{
				input: "PRIVATE_KEY = $GOPATH:invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid JSON input",
			args: args{
				input: `VERSION <= [["npm", "version"]`,
			},
			wantErr: true,
		},
		{
			name: "invalid operation",
			args: args{
				input: `VERSION <> 123`,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := variable.Parse(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
