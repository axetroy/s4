package configuration_test

import (
	"github.com/axetroy/s4/core/configuration"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		wantC   *configuration.Configuration
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				content: []byte(`
# run this config file with [s4]
CONNECT axetroy@192.168.0.1:22

ENV PRIVATE_KEY = 123
ENV TOKEN = xxxx

CD /root # execute the root directory of the script

COPY ./README ./test

RUN ls -lh

RUN echo "hello    world"
`),
			},
			wantC: &configuration.Configuration{
				Host:     "192.168.0.1",
				Port:     "22",
				Username: "axetroy",
				CWD:      "",
				Env: map[string]string{
					"PRIVATE_KEY": "123",
					"TOKEN":       "xxxx",
				},
				Var: map[string]string{},
				Actions: []configuration.Action{
					{
						Action:    "CD",
						Arguments: []string{"/root"},
					},
					{
						Action:    "COPY",
						Arguments: []string{"./README", "./test"},
					},
					{
						Action:    "RUN",
						Arguments: []string{"ls", "-lh"},
					},
					{
						Action:    "RUN",
						Arguments: []string{"echo", "\"hello", "", "", "", "world\""},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "parse CMD",
			args: args{
				content: []byte(`CMD ["ls", "-lh"]`),
			},
			wantC: &configuration.Configuration{
				Env: map[string]string{},
				Var: map[string]string{},
				Actions: []configuration.Action{
					{
						Action:    "CMD",
						Arguments: []string{"ls", "-lh"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotC, err := configuration.Parse(tt.args.content)
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
