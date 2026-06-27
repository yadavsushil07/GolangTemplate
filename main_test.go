package main

import "testing"

func TestOTPFlowAndRateLimit(t *testing.T) {
	app := &app{}
	app.init()

	otp, err := app.requestOTP("demo@example.com")
	if err != nil {
		t.Fatalf("requestOTP failed: %v", err)
	}
	if len(otp) != 6 {
		t.Fatalf("expected 6 digit otp, got %q", otp)
	}

	token, err := app.verifyOTP("demo@example.com", otp)
	if err != nil {
		t.Fatalf("verifyOTP failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected jwt token")
	}

	if !app.rateLimit("10.0.0.1") {
		t.Fatal("expected first request to be allowed")
	}
	for i := 0; i < 5; i++ {
		app.rateLimit("10.0.0.1")
	}
	if app.rateLimit("10.0.0.1") {
		t.Fatal("expected rate limit to block excess requests")
	}
}
