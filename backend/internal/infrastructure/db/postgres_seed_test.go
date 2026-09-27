package db

import "testing"

func TestDefaultSeedMeterProfile(t *testing.T) {
	provider, region, rateType := defaultSeedMeterProfile()
	if provider != "bia" {
		t.Fatalf("provider = %q, want %q", provider, "bia")
	}
	if region != "CUND-EAST" {
		t.Fatalf("region = %q, want %q", region, "CUND-EAST")
	}
	if rateType != "industrial" {
		t.Fatalf("rate_type = %q, want %q", rateType, "industrial")
	}
}
