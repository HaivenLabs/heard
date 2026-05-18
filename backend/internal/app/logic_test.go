package app

import "testing"

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
