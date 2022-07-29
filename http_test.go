package main

import "testing"

func Test_getAppDomain(t *testing.T) {
	tests := []struct {
		ip   string
		want string
	}{
		{"", ""},
		{"localhost", ""},

		{"1.1.1.1", "1-1-1-1.app.rethinkraw.com"},
		{"127.0.0.1", "127-0-0-1.app.rethinkraw.com"},
		{"192.168.1.1", "192-168-1-1.app.rethinkraw.com"},

		{"::", "0--0.app.rethinkraw.com"},
		{"::1", "0--1.app.rethinkraw.com"},
		{"2001:1::1", "2001-1--1.app.rethinkraw.com"},
		{"::ffff:0:0", "0-0-0-0.app.rethinkraw.com"},
	}
	for _, tt := range tests {
		if got := getAppDomain(tt.ip); got != tt.want {
			t.Errorf("getAppDomain(%q) = %q, want %q", tt.ip, got, tt.want)
		}
	}
}
