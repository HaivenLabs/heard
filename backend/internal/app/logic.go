package app

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode"
)

var errValidation = errors.New("validation failed")

func validationError(message string) error {
	return fmt.Errorf("%w: %s", errValidation, message)
}

func validateContactDetails(email, phone string, required bool) error {
	email = strings.TrimSpace(email)
	phone = strings.TrimSpace(phone)
	if required && email == "" && phone == "" {
		return validationError("a phone number or email is required")
	}
	if email != "" {
		parsed, err := mail.ParseAddress(email)
		if err != nil || !strings.EqualFold(parsed.Address, email) || len(email) > 254 {
			return validationError("enter a valid email address")
		}
	}
	if phone != "" && !isValidPhoneFormat(phone) {
		return validationError("enter a valid phone number")
	}
	return nil
}

func isValidPhoneFormat(phone string) bool {
	digits := 0
	for index, character := range phone {
		switch {
		case unicode.IsDigit(character):
			digits++
		case strings.ContainsRune(" ()-.", character):
		case character == '+' && index == 0:
		default:
			return false
		}
	}
	return digits >= 10 && digits <= 15 && len(phone) <= 32
}

func validateMarketingLeadRequest(req createMarketingLeadRequest) error {
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.WorkEmail)
	restaurantName := strings.TrimSpace(req.RestaurantName)
	challenge := strings.TrimSpace(req.Challenge)

	if len(name) < 2 || len(name) > 120 {
		return validationError("enter your name")
	}
	if email == "" {
		return validationError("work email is required")
	}
	if err := validateContactDetails(email, req.Phone, false); err != nil {
		return err
	}
	if len(restaurantName) < 2 || len(restaurantName) > 160 {
		return validationError("enter your restaurant or brand name")
	}
	switch req.LocationCount {
	case "1", "2-4", "5-19", "20-49", "50+":
	default:
		return validationError("choose the number of restaurant locations")
	}
	if len(challenge) > 1000 {
		return validationError("keep your note under 1000 characters")
	}
	switch req.Source {
	case "marketing_site", "guest_demo":
	default:
		return validationError("invalid contact request source")
	}
	if !req.ContactConsent {
		return validationError("contact consent is required")
	}
	return nil
}

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
