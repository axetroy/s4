package lib_test

import (
	"github.com/axetroy/s4/lib"
	"reflect"
	"testing"
)

func TestFileParser(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name    string
		args    args
		want    *lib.Files
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				source: `./start.py ./server`,
			},
			want: &lib.Files{
				Source:      []string{"./start.py"},
				Destination: "./server",
			},
		},
		{
			name: "multiple",
			args: args{
				source: `./start.py ./stop.py ./server`,
			},
			want: &lib.Files{
				Source:      []string{"./start.py", "./stop.py"},
				Destination: "./server",
			},
		},
		{
			name: "single",
			args: args{
				source: `./start.py`,
			},
			want: &lib.Files{
				Source:      []string{},
				Destination: "./start.py",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lib.FileParser(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileParser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileParser() = %v, want %v", got, tt.want)
			}
		})
	}
}
