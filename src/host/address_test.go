package host_test

import (
	"github.com/axetroy/s4/src/host"
	"reflect"
	"testing"
)

func TestParseAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		want    host.Address
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				address: "root@192.168.0.1:22",
			},
			want: host.Address{
				Host:     "192.168.0.1",
				Port:     "22",
				Username: "root",
			},
		},
		{
			name: "invalid-1",
			args: args{
				address: "192.168.0.1:22",
			},
			wantErr: true,
		},
		{
			name: "invalid-2",
			args: args{
				address: "192.168.0.1",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			args: args{
				address: "root@192.168.0.1:abc",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := host.Parse(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
