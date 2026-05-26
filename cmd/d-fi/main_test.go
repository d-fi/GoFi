package main

import "testing"

func TestShouldPauseOnError(t *testing.T) {
	tests := []struct {
		name string
		goos string
		env  string
		want bool
	}{
		{name: "windows", goos: "windows", want: true},
		{name: "non windows", goos: "linux", want: false},
		{name: "disabled with one", goos: "windows", env: "1", want: false},
		{name: "disabled with true", goos: "windows", env: "true", want: false},
		{name: "disabled with yes", goos: "windows", env: "yes", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldPauseOnError(tt.goos, func(key string) string {
				if key == "DFI_NO_PAUSE" {
					return tt.env
				}
				return ""
			})
			if got != tt.want {
				t.Fatalf("shouldPauseOnError() = %v, want %v", got, tt.want)
			}
		})
	}
}
