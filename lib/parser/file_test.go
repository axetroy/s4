package parser

import (
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
		want    *Files
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				source: `./start.py ./server`,
			},
			want: &Files{
				Source:      []string{"./start.py"},
				Destination: "./server",
			},
		},
		{
			name: "multiple",
			args: args{
				source: `./start.py ./stop.py ./server`,
			},
			want: &Files{
				Source:      []string{"./start.py", "./stop.py"},
				Destination: "./server",
			},
		},
		{
			name: "single",
			args: args{
				source: `./start.py`,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FileParser(tt.args.source)
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
