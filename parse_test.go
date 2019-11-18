package gpsnmea

import "testing"

func TestDecimalDegreeToLat(t *testing.T) {
	type args struct {
		lat float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			args: args{-75.60176188566656},
			want: "0609.89000,N",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecimalDegreeToLat(tt.args.lat); got != tt.want {
				t.Errorf("DecimalDegreeToLat() = %v, want %v", got, tt.want)
			}
		})
	}
}
