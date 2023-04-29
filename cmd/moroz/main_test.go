package main

import (
	"testing"
)

func Test_validateConfigExists(t *testing.T) {
	type args struct {
		configsPath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "config dir is missing",
			args: args{
				configsPath: "./fred",
			},
			want: false,
		},
		{
			name: "config dir exists",
			args: args{
				configsPath: "/tmp",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateConfigExists(tt.args.configsPath); got != tt.want {
				t.Errorf("validateConfigExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
