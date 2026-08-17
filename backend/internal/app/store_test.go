package app

import "testing"

func TestDemoReviewDestinations(t *testing.T) {
	if demoGoogleReviewURL != "https://maps.app.goo.gl/D3cEeXBEtGaKF2Lz8" {
		t.Fatalf("unexpected demo Google review URL: %q", demoGoogleReviewURL)
	}
	if demoYelpReviewURL != "https://www.yelp.com/biz/nom-san-juan-capistrano" {
		t.Fatalf("unexpected demo Yelp review URL: %q", demoYelpReviewURL)
	}
}
