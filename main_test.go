package main

import "testing"

func TestParseHIDIdleTime(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    int64
		wantErr bool
	}{
		{
			name: "typical ioreg output",
			output: `+-o IOHIDSystem  <class IOHIDSystem, id 0x100000abc, registered, matched, active, busy 0 (0 ms), retain 20>
    {
      "IOClass" = "IOHIDSystem"
      "HIDIdleTime" = 12345678901
      "IOProviderClass" = "IOResources"
    }`,
			want: 12345678901,
		},
		{
			name:   "zero idle time",
			output: `"HIDIdleTime" = 0`,
			want:   0,
		},
		{
			name:    "missing key",
			output:  `"IOClass" = "IOHIDSystem"`,
			wantErr: true,
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHIDIdleTime(tt.output)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIdleNanosToSeconds(t *testing.T) {
	tests := []struct {
		nanos int64
		want  int64
	}{
		{0, 0},
		{999999999, 0},
		{1000000000, 1},
		{12345678901, 12},
		{300000000000, 300},
	}
	for _, tt := range tests {
		if got := idleNanosToSeconds(tt.nanos); got != tt.want {
			t.Errorf("idleNanosToSeconds(%d) = %d, want %d", tt.nanos, got, tt.want)
		}
	}
}

func TestIsUserActive(t *testing.T) {
	tests := []struct {
		idleSeconds int64
		threshold   int
		want        bool
	}{
		{0, 300, true},
		{299, 300, true},
		{300, 300, false},
		{301, 300, false},
		{5, 10, true},
		{10, 10, false},
	}
	for _, tt := range tests {
		if got := isUserActive(tt.idleSeconds, tt.threshold); got != tt.want {
			t.Errorf("isUserActive(%d, %d) = %v, want %v", tt.idleSeconds, tt.threshold, got, tt.want)
		}
	}
}

func TestGetIdleThreshold(t *testing.T) {
	tests := []struct {
		name string
		cfg  config
		want int
	}{
		{"default when unset", config{}, 300},
		{"default when zero", config{IdleThreshold: 0}, 300},
		{"default when negative", config{IdleThreshold: -5}, 300},
		{"custom value", config{IdleThreshold: 60}, 60},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.getIdleThreshold(); got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"v2.0.8", "v2.0.9", true},
		{"v2.0.8", "v2.0.10", true}, // two-digit patch must not sort as a string
		{"v2.0.8", "v2.0.11", true},
		{"v2.0.9", "v2.0.10", true},
		{"v2.0.10", "v2.0.9", false},
		{"v2.0.10", "v2.0.10", false},
		{"v2.0.8", "v2.1.0", true},
		{"v2.1.0", "v2.0.99", false},
		{"v2.0.8", "v3.0.0", true},
		{"v3.0.0", "v2.9.9", false},
		{"2.0.8", "v2.0.10", true}, // prefix optional on either side
		{"v2.0.8", "2.0.10", true},
		{"dev", "v2.0.10", true},
		{"v2.0.8", "garbage", false}, // unparseable release tag is never newer
		{"v2.0.8", "", false},
	}
	for _, tt := range tests {
		if got := isNewerVersion(tt.current, tt.latest); got != tt.want {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}
