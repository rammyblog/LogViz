package logviz

import (
	"net/http"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	// Test case 1: X-Real-IP header set
	req1 := &http.Request{
		Header:     http.Header{"X-Real-IP": []string{"192.168.1.1"}},
		RemoteAddr: "127.0.0.1:12345",
	}
	if ip := getClientIP(req1); ip != "192.168.1.1" {
		t.Errorf("Test case 1 failed: expected IP %s, got %s", "192.168.1.1", ip)
	}

	// Test case 2: X-Forwarded-For header set
	req2 := &http.Request{
		Header:     http.Header{"X-Forwarded-For": []string{"192.168.1.2, 192.168.1.3"}},
		RemoteAddr: "127.0.0.1:12345",
	}
	if ip := getClientIP(req2); ip != "192.168.1.2" {
		t.Errorf("Test case 2 failed: expected IP %s, got %s", "192.168.1.2", ip)
	}

	// Test case 3: Neither X-Real-IP nor X-Forwarded-For header set
	req3 := &http.Request{
		RemoteAddr: "127.0.0.1:12345",
	}
	if ip := getClientIP(req3); ip != "127.0.0.1" {
		t.Errorf("Test case 3 failed: expected IP %s, got %s", "127.0.0.1", ip)
	}
}
