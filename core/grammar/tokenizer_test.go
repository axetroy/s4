package grammar_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/axetroy/s4/core/grammar"
	"github.com/axetroy/s4/core/host"
)

func TestTokenizer(t *testing.T) {
	type args struct {
		input string
	}

	password := "123123"
	privateKeyFile := "./path/to/private/key/file"

	tests := []struct {
		name    string
		args    args
		want    []grammar.Token
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				input: `RUN 192.168.0.1
MOVE data.db data.db.bak
COPY data.db data.db.bak
DELETE file1.txt file2.txt
		`,
			},
			want: []grammar.Token{
				{
					Key: "RUN",
					Node: grammar.NodeRun{
						Commands: []grammar.NodeRunCommand{
							{
								Command:    []string{"192.168.0.1"},
								RunInLocal: false,
								SourceCode: "192.168.0.1",
							},
						},
						SourceCode: "192.168.0.1",
					},
				},
				{
					Key: "MOVE",
					Node: grammar.NodeCopy{
						Source:      "data.db",
						Destination: "data.db.bak",
						SourceCode:  "data.db data.db.bak",
					},
				},
				{
					Key: "COPY",
					Node: grammar.NodeCopy{
						Source:      "data.db",
						Destination: "data.db.bak",
						SourceCode:  "data.db data.db.bak",
					},
				},
				{
					Key: "DELETE",
					Node: grammar.NodeDelete{
						Targets:    []string{"file1.txt", "file2.txt"},
						SourceCode: "file1.txt file2.txt",
					},
				},
			},
		},
		{
			name: "basic with multiple blank",
			args: args{
				input: "RUN     192.168.0.1",
			},
			want: []grammar.Token{
				{
					Key: "RUN",
					Node: grammar.NodeRun{
						Commands: []grammar.NodeRunCommand{
							{
								Command:    []string{"192.168.0.1"},
								RunInLocal: false,
								SourceCode: "192.168.0.1",
							},
						},
						SourceCode: "192.168.0.1",
					},
				},
			},
		},
		{
			name: "basic with comment single line",
			args: args{
				input: `# server host address
RUN     192.168.0.1`,
			},
			want: []grammar.Token{
				{
					Key: "RUN",
					Node: grammar.NodeRun{
						Commands: []grammar.NodeRunCommand{
							{
								Command:    []string{"192.168.0.1"},
								RunInLocal: false,
								SourceCode: "192.168.0.1",
							},
						},
						SourceCode: "192.168.0.1",
					},
				},
			},
		},
		{
			name: "basic with tail comment",
			args: args{
				input: `RUN 192.168.0.1 # set your server address`,
			},
			want: []grammar.Token{
				{
					Key: "RUN",
					Node: grammar.NodeRun{
						Commands: []grammar.NodeRunCommand{
							{
								Command:    []string{"192.168.0.1"},
								RunInLocal: false,
								SourceCode: "192.168.0.1",
							},
						},
						SourceCode: "192.168.0.1",
					},
				},
			},
		},
		{
			name: "basic with prefix comment",
			args: args{
				input: `# HOST 192.168.0.1`,
			},
			want: []grammar.Token{},
		},
		{
			name: "multiple field",
			args: args{
				input: `CONNECT axetroy@192.168.0.1:22
RUN ls -lh

		`,
			},
			want: []grammar.Token{
				{
					Key: "CONNECT",
					Node: grammar.NodeConnect{
						Host:       "192.168.0.1",
						Port:       "22",
						Username:   "axetroy",
						SourceCode: "axetroy@192.168.0.1:22",
					},
				},
				{
					Key: "RUN",
					Node: grammar.NodeRun{
						Commands: []grammar.NodeRunCommand{
							{
								Command:    []string{"ls -lh"},
								RunInLocal: false,
								SourceCode: "ls -lh",
							},
						},
						SourceCode: "ls -lh",
					},
				},
			},
		},
		{
			name: "multiple values",
			args: args{
				input: `UPLOAD ./README.md ./start.py ./dist`,
			},
			want: []grammar.Token{
				{
					Key: "UPLOAD",
					Node: grammar.NodeUpload{
						SourceFiles:    []string{"./README.md", "./start.py"},
						DestinationDir: "./dist",
						SourceCode:     "./README.md ./start.py ./dist",
					},
				},
			},
		},
		{
			name: "invalid keyword",
			args: args{
				input: "INVALID KEYWORD",
			},
			want:    []grammar.Token{},
			wantErr: true,
		},
		{
			name: "connect with password",
			args: args{
				input: "CONNECT root@192.168.0.1:2222 WITH PASSWORD 123123",
			},
			want: []grammar.Token{
				{
					Key: grammar.ActionCONNECT,
					Node: grammar.NodeConnect{
						Host:        "192.168.0.1",
						Username:    "root",
						Port:        "2222",
						ConnectType: &host.ConnectTypePassword,
						Password:    &password,
						SourceCode:  "root@192.168.0.1:2222 WITH PASSWORD 123123",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "connect with private file",
			args: args{
				input: "CONNECT root@192.168.0.1:2222 WITH FILE ./path/to/private/key/file",
			},
			want: []grammar.Token{
				{
					Key: grammar.ActionCONNECT,
					Node: grammar.NodeConnect{
						Host:        "192.168.0.1",
						Username:    "root",
						Port:        "2222",
						ConnectType: &host.ConnectTypePrivateKeyFile,
						Password:    &privateKeyFile,
						SourceCode:  "root@192.168.0.1:2222 WITH FILE ./path/to/private/key/file",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid keyword with comment",
			args: args{
				input: "# INVALID KEYWORD",
			},
			want:    []grammar.Token{},
			wantErr: false,
		},
		{
			name: "Invalid grammar",
			args: args{
				input: "(abc)",
			},
			want:    []grammar.Token{},
			wantErr: true,
		},
		{
			name: "Empty value",
			args: args{
				input: "CONNECT",
			},
			want:    []grammar.Token{},
			wantErr: true,
		},
		{
			name: "Invalid ENV",
			args: args{
				input: "ENV PRIVATE_KEY",
			},
			want:    []grammar.Token{},
			wantErr: true,
		},
		{
			name: "Invalid ENV",
			args: args{
				input: "ENV PRIVATE_KEY = ",
			},
			want:    []grammar.Token{},
			wantErr: true,
		},
		{
			name: "parse ENV",
			args: args{
				input: "ENV PRIVATE_KEY = xxx",
			},
			want: []grammar.Token{
				{
					Key: "ENV",
					Node: grammar.NodeEnv{
						Key:        "PRIVATE_KEY",
						Value:      "xxx",
						SourceCode: "PRIVATE_KEY = xxx",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multi line script for RUN",
			args: args{
				input: `
RUN yarn \
	&& npm run build \
	&& env
		`,
			},
			want: []grammar.Token{
				{
					Key: "RUN",
					Node: grammar.NodeRun{
						Commands: []grammar.NodeRunCommand{
							{
								Command:    []string{"yarn", "npm run build", "env"},
								RunInLocal: false,
								SourceCode: "yarn  && npm run build  && env",
							},
						},
						SourceCode: `yarn  && npm run build  && env`,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multi line script for RUN with tail space blank",
			args: args{
				input: `RUN yarn \
&& env
`,
			},
			want: []grammar.Token{
				{
					Key: "RUN",
					Node: grammar.NodeRun{
						Commands: []grammar.NodeRunCommand{
							{
								Command:    []string{`yarn`, `env`},
								RunInLocal: false,
								SourceCode: `yarn && env`,
							},
						},
						SourceCode: `yarn && env`,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "parse var literal",
			args: args{
				input: `
		VAR name = axetroy
		`,
			},
			want: []grammar.Token{
				{
					Key: "VAR",
					Node: grammar.NodeVar{
						Key:        "name",
						Literal:    &grammar.NodeVarLiteral{Value: "axetroy"},
						SourceCode: "name = axetroy",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "parse var env",
			args: args{
				input: `
		VAR remote_home = $HOME:remote
		VAR local_home = $HOME:local
		`,
			},
			want: []grammar.Token{
				{
					Key: "VAR",
					Node: grammar.NodeVar{
						Key: "remote_home",
						Env: &grammar.NodeVarEnv{
							Local: false,
							Key:   "HOME",
						},
						SourceCode: "remote_home = $HOME:remote",
					},
				},
				{
					Key: "VAR",
					Node: grammar.NodeVar{
						Key: "local_home",
						Env: &grammar.NodeVarEnv{
							Local: true,
							Key:   "HOME",
						},
						SourceCode: "local_home = $HOME:local",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "parse var command",
			args: args{
				input: `
		VAR local_home <= ["echo", "$HOME"]
		VAR remote_home <= echo $HOME
		`,
			},
			want: []grammar.Token{
				{
					Key: "VAR",
					Node: grammar.NodeVar{
						Key: "local_home",
						Command: &grammar.NodeVarCommand{
							Local:   true,
							Command: []string{"echo", "$HOME"},
						},
						SourceCode: `local_home <= ["echo", "$HOME"]`,
					},
				},
				{
					Key: "VAR",
					Node: grammar.NodeVar{
						Key: "remote_home",
						Command: &grammar.NodeVarCommand{
							Local:   false,
							Command: []string{"echo", "$HOME"},
						},
						SourceCode: `remote_home <= echo $HOME`,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "run local command with RUN",
			args: args{
				input: `
		RUN ["npm", "run", "build"]
		`,
			},
			want: []grammar.Token{
				{
					Key: "RUN",
					Node: grammar.NodeRun{
						Commands: []grammar.NodeRunCommand{
							{
								Command:    []string{"npm", "run", "build"},
								RunInLocal: true,
								SourceCode: `["npm", "run", "build"]`,
							},
						},
						SourceCode: `["npm", "run", "build"]`,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.Tokenizer(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tokenizer() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				d1, _ := json.Marshal(got)
				d2, _ := json.Marshal(tt.want)
				fmt.Println(string(d1))
				fmt.Println(string(d2))
				t.Errorf("Tokenizer() = \nresult: %+v\nexpect: %+v", got, tt.want)
			}
		})
	}
}
