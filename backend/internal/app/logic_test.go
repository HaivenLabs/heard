package app

import "testing"

func TestNullableUUID(t *testing.T) {
	value, err := nullableUUID("11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("parse UUID: %v", err)
	}
	if value == nil || value.String() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("got unexpected UUID %v", value)
	}

	empty, err := nullableUUID("")
	if err != nil {
		t.Fatalf("empty UUID: %v", err)
	}
	if empty != nil {
		t.Fatalf("expected nil UUID, got %v", empty)
	}
}

func TestValidateContactDetailsAcceptsFormattedEmailOrPhone(t *testing.T) {
	if err := validateContactDetails("guest@example.com", "", true); err != nil {
		t.Fatalf("valid email rejected: %v", err)
	}
	if err := validateContactDetails("", "+1 (206) 555-0142", true); err != nil {
		t.Fatalf("valid phone rejected: %v", err)
	}
}

func TestValidateContactDetailsRejectsMissingGiveawayContact(t *testing.T) {
	if err := validateContactDetails("", "", true); err == nil {
		t.Fatal("expected a contact method to be required")
	}
}

func TestValidateContactDetailsRejectsInvalidFormats(t *testing.T) {
	if err := validateContactDetails("not-an-email", "", false); err == nil {
		t.Fatal("expected invalid email to be rejected")
	}
	if err := validateContactDetails("", "call-me-maybe", false); err == nil {
		t.Fatal("expected invalid phone to be rejected")
	}
	if err := validateContactDetails("", "555-12", false); err == nil {
		t.Fatal("expected short phone to be rejected")
	}
}

func TestValidateMarketingLeadRequestAcceptsRestaurantProspect(t *testing.T) {
	req := createMarketingLeadRequest{
		Name:           "Avery Chen",
		WorkEmail:      "avery@restaurant.example",
		RestaurantName: "Cedar & Salt",
		LocationCount:  "2-4",
		Source:         "guest_demo",
		ContactConsent: true,
	}

	if err := validateMarketingLeadRequest(req); err != nil {
		t.Fatalf("valid marketing lead rejected: %v", err)
	}
}

func TestValidateMarketingLeadRequestRejectsInvalidInput(t *testing.T) {
	valid := createMarketingLeadRequest{
		Name:           "Avery Chen",
		WorkEmail:      "avery@restaurant.example",
		RestaurantName: "Cedar & Salt",
		LocationCount:  "2-4",
		Source:         "marketing_site",
		ContactConsent: true,
	}

	tests := map[string]createMarketingLeadRequest{
		"missing consent":    func() createMarketingLeadRequest { item := valid; item.ContactConsent = false; return item }(),
		"invalid email":      func() createMarketingLeadRequest { item := valid; item.WorkEmail = "avery"; return item }(),
		"invalid phone":      func() createMarketingLeadRequest { item := valid; item.Phone = "call me"; return item }(),
		"unknown locations":  func() createMarketingLeadRequest { item := valid; item.LocationCount = "many"; return item }(),
		"unknown source":     func() createMarketingLeadRequest { item := valid; item.Source = "mystery"; return item }(),
		"missing restaurant": func() createMarketingLeadRequest { item := valid; item.RestaurantName = ""; return item }(),
	}

	for name, req := range tests {
		t.Run(name, func(t *testing.T) {
			if err := validateMarketingLeadRequest(req); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestSentimentFromRating(t *testing.T) {
	cases := map[int]string{
		1: "negative",
		2: "negative",
		3: "neutral",
		4: "positive",
		5: "positive",
	}

	for rating, want := range cases {
		if got := sentimentFromRating(rating); got != want {
			t.Fatalf("rating %d: got %q want %q", rating, got, want)
		}
	}
}

func TestShouldCreateRecoveryCase(t *testing.T) {
	if !shouldCreateRecoveryCase(1, "") {
		t.Fatal("expected rating 1 to create a recovery case")
	}
	if !shouldCreateRecoveryCase(2, "negative") {
		t.Fatal("expected negative sentiment to create a recovery case")
	}
	if shouldCreateRecoveryCase(4, "") {
		t.Fatal("expected rating 4 to skip recovery case creation")
	}
}

func TestFlyerGiveawayCreatesRecoveryCaseBelowFive(t *testing.T) {
	response := FeedbackResponse{
		Rating:    4,
		Sentiment: "positive",
		Metadata: map[string]any{
			"campaign_type":        "flyer_giveaway",
			"follow_up_required":   true,
			"public_review_prompt": false,
		},
	}

	if !shouldCreateRecoveryCaseForResponse(response) {
		t.Fatal("expected flyer giveaway rating below 5 to create a recovery case")
	}
	if got := recoveryReasonForResponse(response); got != "flyer_giveaway_under_five_follow_up" {
		t.Fatalf("got reason %q", got)
	}
}

func TestFlyerGiveawayFiveSkipsRecoveryCase(t *testing.T) {
	response := FeedbackResponse{
		Rating:    5,
		Sentiment: "positive",
		Metadata: map[string]any{
			"campaign_type":        "flyer_giveaway",
			"public_review_prompt": true,
			"follow_up_required":   false,
		},
	}

	if shouldCreateRecoveryCaseForResponse(response) {
		t.Fatal("expected flyer giveaway rating 5 to skip recovery case creation")
	}
}

func TestPreviewComment(t *testing.T) {
	short := "Warm service, cold fries."
	if got := previewComment(short); got != short {
		t.Fatalf("short preview changed: %q", got)
	}

	long := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"
	if got := previewComment(long); len(got) != 140 {
		t.Fatalf("expected 140-char preview, got %d", len(got))
	}
}
