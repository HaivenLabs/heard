package app

import (
	"fmt"
	"strings"
	"testing"
)

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
	valid := []struct {
		email string
		phone string
	}{
		{email: "guest@example.com"},
		{email: "first.last+takeout@restaurant.co.uk"},
		{phone: "(206) 555-0142"},
		{phone: "1-206-555-0142"},
		{phone: "+1 (206) 555-0142"},
		{phone: "+44 20 7946 0958"},
	}
	for _, item := range valid {
		if err := validateContactDetails(item.email, item.phone, true); err != nil {
			t.Fatalf("valid contact rejected (%q, %q): %v", item.email, item.phone, err)
		}
	}
}

func TestValidateContactDetailsRejectsMissingGiveawayContact(t *testing.T) {
	if err := validateContactDetails("", "", true); err == nil {
		t.Fatal("expected a contact method to be required")
	}
}

func TestValidateContactDetailsRejectsInvalidFormats(t *testing.T) {
	invalid := []struct {
		email string
		phone string
	}{
		{email: "not-an-email"},
		{email: "guest@example.c"},
		{email: ".guest@example.com"},
		{email: "guest..name@example.com"},
		{email: "guest@example..com"},
		{email: "guest@-example.com"},
		{email: "guest@example_.com"},
		{phone: "call-me-maybe"},
		{phone: "555-12"},
		{phone: "000-000-0000"},
		{phone: "123-456-7890"},
		{phone: "206-155-0142"},
		{phone: "+01234567890"},
		{phone: "+1 206 555 014"},
		{phone: "206+555+0142"},
		{phone: "(206 555-0142"},
		{phone: "-206-555-0142"},
		{phone: "206-555-0142-"},
	}
	for _, item := range invalid {
		if err := validateContactDetails(item.email, item.phone, false); err == nil {
			t.Fatalf("expected invalid contact to be rejected (%q, %q)", item.email, item.phone)
		}
	}
}

func TestValidateContactDetailsAllowsOptionalContactForNonGiveaway(t *testing.T) {
	if err := validateContactDetails("", "", false); err != nil {
		t.Fatalf("optional contact rejected: %v", err)
	}
}

func TestValidateContactDetailsRequiresOnlyOneGiveawayContactMethod(t *testing.T) {
	for _, item := range []struct {
		name  string
		email string
		phone string
	}{
		{name: "email", email: "guest@example.com"},
		{name: "phone", phone: "(206) 555-0142"},
	} {
		t.Run(item.name, func(t *testing.T) {
			if err := validateContactDetails(item.email, item.phone, true); err != nil {
				t.Fatalf("valid giveaway contact rejected: %v", err)
			}
		})
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

func TestNormalizeRatingFaceSet(t *testing.T) {
	for _, value := range []string{"heard", "clay", "glass", "minimal", "retro"} {
		got, err := normalizeRatingFaceSet(value)
		if err != nil || got != value {
			t.Fatalf("normalize %q: got %q err=%v", value, got, err)
		}
	}
	got, err := normalizeRatingFaceSet("")
	if err != nil || got != "heard" {
		t.Fatalf("empty set should default to heard: got %q err=%v", got, err)
	}
	if _, err := normalizeRatingFaceSet("ios"); err == nil {
		t.Fatal("expected unsupported rating face set to fail")
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

func TestRecoveryStatusTransitionMatrix(t *testing.T) {
	allowed := [][2]string{{"new", "open"}, {"open", "assigned"}, {"assigned", "resolved"}, {"open", "closed"}}
	for _, transition := range allowed {
		if !recoveryStatusTransitionAllowed(transition[0], transition[1]) {
			t.Fatalf("expected %s -> %s to be allowed", transition[0], transition[1])
		}
	}
	for _, transition := range [][2]string{{"resolved", "open"}, {"closed", "assigned"}, {"spam", "open"}, {"new", "resolved"}} {
		if recoveryStatusTransitionAllowed(transition[0], transition[1]) {
			t.Fatalf("expected %s -> %s to be rejected", transition[0], transition[1])
		}
	}
}

func TestServerCampaignMetadataCannotOptOutOfFlyerRules(t *testing.T) {
	metadata := enforceCampaignMetadata(map[string]any{"campaign_type": "other", "follow_up_required": false, "public_review_prompt": true}, true, 4)
	if metadata["campaign_type"] != "flyer_giveaway" || metadata["follow_up_required"] != true || metadata["public_review_prompt"] != false {
		t.Fatalf("server rules were not enforced: %#v", metadata)
	}
}

func TestValidatePublicFeedbackInputBounds(t *testing.T) {
	if err := validateFeedbackSessionInput(createFeedbackSessionRequest{Token: "token", GuestName: strings.Repeat("x", 121)}); err == nil {
		t.Fatal("expected oversized guest name to be rejected")
	}
	if err := validateSubmitFeedbackInput(submitFeedbackRequest{FeedbackSessionID: "11111111-1111-1111-1111-111111111111", Rating: 4, Comment: strings.Repeat("x", 2001)}); err == nil {
		t.Fatal("expected oversized feedback comment to be rejected")
	}
	if err := validateSubmitFeedbackInput(submitFeedbackRequest{FeedbackSessionID: "11111111-1111-1111-1111-111111111111", Rating: 4, Categories: make([]string, 11)}); err == nil {
		t.Fatal("expected oversized category list to be rejected")
	}
	metadata := map[string]any{}
	for index := 0; index < 21; index++ {
		metadata[fmt.Sprintf("key-%02d", index)] = true
	}
	if err := validatePublicMetadata(metadata); err == nil {
		t.Fatal("expected excessive metadata keys to be rejected")
	}
}
