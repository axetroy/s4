package grammar_test

import (
	"github.com/axetroy/s4/core/grammar"
	"reflect"
	"testing"
)

func TestTokenizer(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    []grammar.Token
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				input: "RUN 192.168.0.1",
			},
			want: []grammar.Token{
				{
					Key:   "RUN",
					Value: []string{"192.168.0.1"},
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
					Key:   "RUN",
					Value: []string{"192.168.0.1"},
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
					Key:   "RUN",
					Value: []string{"192.168.0.1"},
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
					Key:   "RUN",
					Value: []string{"192.168.0.1"},
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
					Key:   "CONNECT",
					Value: []string{"axetroy@192.168.0.1:22"},
				},
				{
					Key:   "RUN",
					Value: []string{"ls", "-lh"},
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
					Key:   "UPLOAD",
					Value: []string{"./README.md", "./start.py", "./dist"},
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
					Key:   "ENV",
					Value: []string{"PRIVATE_KEY", "xxx"},
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
					Key:   "RUN",
					Value: []string{"yarn", "&&", "npm", "run", "build", "&&", "env"},
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
					Key:   "RUN",
					Value: []string{"yarn", "&&", "env"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.Tokenizer(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tokenizer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tokenizer() = %v, want %v", got, tt.want)
			}
		})
	}
}
