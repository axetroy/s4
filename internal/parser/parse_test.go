package parser

import (
	"reflect"
	"testing"
)

func TestRemoveComment(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			want: "",
			args: args{
				value: `# hello world`,
			},
		},
		{
			name: "basic",
			want: "PORT 22",
			args: args{
				value: `PORT 22 # remote server port`,
			},
		},
		{
			name: "basic",
			want: "PORT 22",
			args: args{
				value: `PORT 22 # remote server port`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveComment(tt.args.value); got != tt.want {
				t.Errorf("RemoveComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		wantC   Config
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				content: []byte(`
# run this config file with [s4]
HOST 192.168.0.1
PORT 22 # remote ssh server port
USERNAME axetroy

CWD /root # execute the root directory of the script

COPY ./README ./test

RUN ls -lh
`),
			},
			wantC: Config{
				Host:     "192.168.0.1",
				Port:     22,
				Username: "axetroy",
				CWD:      "/root",
				Actions: []Action{
					{
						Action:    "CWD",
						Arguments: "/root",
					},
					{
						Action:    "COPY",
						Arguments: "./README ./test",
					},
					{
						Action:    "RUN",
						Arguments: "ls -lh",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotC, err := Parse(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("Parse() = %v, want %v", gotC, tt.wantC)
			}
		})
	}
}
