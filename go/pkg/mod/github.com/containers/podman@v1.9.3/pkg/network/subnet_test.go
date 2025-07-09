package network

import (
	"net"
	"reflect"
	"testing"
)

func TestNextSubnet(t *testing.T) {
	type args struct {
		subnet *net.IPNet
	}
	tests := []struct {
		name    string
		args    args
		want    *net.IPNet
		wantErr bool
	}{
		{"class b", args{subnet: parseCIDR("192.168.0.0/16")}, parseCIDR("192.169.0.0/16"), false},
		{"class c", args{subnet: parseCIDR("192.168.1.0/24")}, parseCIDR("192.168.2.0/24"), false},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			got, err := NextSubnet(test.args.subnet)
			if (err != nil) != test.wantErr {
				t.Errorf("NextSubnet() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("NextSubnet() got = %v, want %v", got, test.want)
			}
		})
	}
}
