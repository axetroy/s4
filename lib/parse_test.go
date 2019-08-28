package lib_test

import (
	"github.com/axetroy/s4/lib"
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
		wantC   *lib.Config
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
			wantC: &lib.Config{
				Host:     "192.168.0.1",
				Port:     "22",
				Username: "axetroy",
				CWD:      "",
				Env: map[string]string{
					"PRIVATE_KEY": "123",
					"TOKEN":       "xxxx",
				},
				Actions: []lib.Action{
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
			wantC: &lib.Config{
				Env: map[string]string{},
				Actions: []lib.Action{
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
			gotC, err := lib.Parse(tt.args.content)
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
