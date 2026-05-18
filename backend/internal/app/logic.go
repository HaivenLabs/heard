package app

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func sentimentFromRating(rating int) string {
	switch {
	case rating <= 2:
		return "negative"
	case rating == 3:
		return "neutral"
	default:
		return "positive"
	}
}

func priorityFromSentiment(sentiment string) string {
	if sentiment == "negative" {
		return "high"
	}
	return "normal"
}

func shouldCreateRecoveryCase(rating int, sentiment string) bool {
	if sentiment == "" {
		sentiment = sentimentFromRating(rating)
	}
	return sentiment == "negative"
}

func shouldCreateRecoveryCaseForResponse(response FeedbackResponse) bool {
	if response.Metadata["follow_up_required"] == true {
		return true
	}
	if response.Metadata["campaign_type"] == "flyer_giveaway" && response.Rating < 5 {
		return true
	}
	return shouldCreateRecoveryCase(response.Rating, response.Sentiment)
}

func recoveryReasonForResponse(response FeedbackResponse) string {
	if response.Metadata["campaign_type"] == "flyer_giveaway" && response.Rating < 5 {
		return "flyer_giveaway_under_five_follow_up"
	}
	return "negative_feedback_rule"
}

func previewComment(comment string) string {
	comment = strings.TrimSpace(comment)
	if len(comment) <= 140 {
		return comment
	}
	return comment[:137] + "..."
}

func createOpaqueToken() (string, error) {
	raw := make([]byte, 18)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(raw), "="), nil
}
