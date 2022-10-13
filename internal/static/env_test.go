package static

import "testing"

func Test_getEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		want     string
	}{
		{"ENVSet", "key", "", "value"},
		{"ENVSetAndFallback", "key", "fallback", "value"},
		{"UseFallback", "", "fallback", "fallback"},
		{"Empty", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key != "" {
				t.Setenv(tt.key, "value")
			}
			if got := getEnv(tt.key, tt.fallback); got != tt.want {
				t.Errorf("getEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
