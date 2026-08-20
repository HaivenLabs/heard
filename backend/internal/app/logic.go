package app

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var errValidation = errors.New("validation failed")

var (
	emailLocalPattern     = regexp.MustCompile("^[A-Za-z0-9!#$%&'*+/=?^_{|}~.-]+$")
	domainLabelPattern    = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)
	topLevelDomainPattern = regexp.MustCompile(`^[A-Za-z]{2,63}$`)
)

func validationError(message string) error {
	return fmt.Errorf("%w: %s", errValidation, message)
}

func validateContactDetails(email, phone string, required bool) error {
	email = strings.TrimSpace(email)
	phone = strings.TrimSpace(phone)
	if required && email == "" && phone == "" {
		return validationError("a phone number or email is required")
	}
	if email != "" && !isValidEmailFormat(email) {
		return validationError("enter a valid email address")
	}
	if phone != "" && !isValidPhoneFormat(phone) {
		return validationError("enter a valid phone number")
	}
	return nil
}

func isValidEmailFormat(email string) bool {
	if len(email) > 254 || strings.Count(email, "@") != 1 {
		return false
	}
	local, domain, found := strings.Cut(email, "@")
	if !found || local == "" || len(local) > 64 || strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") || strings.Contains(local, "..") {
		return false
	}
	if !emailLocalPattern.MatchString(local) || domain == "" || len(domain) > 253 || strings.Contains(domain, "..") {
		return false
	}
	labels := strings.Split(domain, ".")
	if len(labels) < 2 || !topLevelDomainPattern.MatchString(labels[len(labels)-1]) {
		return false
	}
	for _, label := range labels {
		if !domainLabelPattern.MatchString(label) {
			return false
		}
	}
	return true
}

func isValidPhoneFormat(phone string) bool {
	if phone == "" || len(phone) > 32 {
		return false
	}
	body := strings.TrimPrefix(phone, "+")
	if body == "" || (body[0] != '(' && (body[0] < '0' || body[0] > '9')) || body[len(body)-1] < '0' || body[len(body)-1] > '9' {
		return false
	}
	digits := strings.Builder{}
	parenthesisDepth := 0
	for index, character := range phone {
		switch {
		case character >= '0' && character <= '9':
			digits.WriteRune(character)
		case strings.ContainsRune(" ()-.", character):
			if character == '(' {
				parenthesisDepth++
			}
			if character == ')' {
				parenthesisDepth--
			}
			if parenthesisDepth < 0 || parenthesisDepth > 1 {
				return false
			}
		case character == '+' && index == 0:
		default:
			return false
		}
	}
	if parenthesisDepth != 0 {
		return false
	}
	number := digits.String()
	if strings.HasPrefix(phone, "+") {
		if len(number) < 8 || len(number) > 15 || number[0] == '0' {
			return false
		}
		if strings.HasPrefix(number, "1") {
			return len(number) == 11 && isValidNANP(number[1:])
		}
		return true
	}
	if len(number) == 11 && strings.HasPrefix(number, "1") {
		number = number[1:]
	}
	return isValidNANP(number)
}

func isValidNANP(number string) bool {
	return len(number) == 10 && number[0] >= '2' && number[0] <= '9' && number[3] >= '2' && number[3] <= '9'
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

func validateFeedbackSessionInput(req createFeedbackSessionRequest) error {
	if length := utf8.RuneCountInString(strings.TrimSpace(req.Token)); length < 1 || length > 128 {
		return validationError("feedback token must be between 1 and 128 characters")
	}
	if utf8.RuneCountInString(strings.TrimSpace(req.Channel)) > 32 {
		return validationError("channel must be 32 characters or fewer")
	}
	if utf8.RuneCountInString(strings.TrimSpace(req.GuestName)) > 120 {
		return validationError("guest name must be 120 characters or fewer")
	}
	if err := validateContactDetails(req.GuestEmail, req.GuestPhone, false); err != nil {
		return err
	}
	return validatePublicMetadata(req.Metadata)
}

func validateSubmitFeedbackInput(req submitFeedbackRequest) error {
	if _, err := uuid.Parse(strings.TrimSpace(req.FeedbackSessionID)); err != nil {
		return validationError("feedback session ID is invalid")
	}
	if req.Rating < 1 || req.Rating > 5 {
		return validationError("rating must be between 1 and 5")
	}
	if utf8.RuneCountInString(strings.TrimSpace(req.Comment)) > 2000 {
		return validationError("feedback comment must be 2000 characters or fewer")
	}
	if utf8.RuneCountInString(strings.TrimSpace(req.GuestName)) > 120 {
		return validationError("guest name must be 120 characters or fewer")
	}
	if len(req.Categories) > 10 {
		return validationError("choose no more than 10 feedback categories")
	}
	seen := map[string]struct{}{}
	for _, raw := range req.Categories {
		category := strings.TrimSpace(raw)
		if length := utf8.RuneCountInString(category); length < 1 || length > 64 {
			return validationError("feedback categories must be between 1 and 64 characters")
		}
		if _, exists := seen[category]; exists {
			return validationError("feedback categories must be unique")
		}
		seen[category] = struct{}{}
	}
	if err := validateContactDetails(req.GuestEmail, req.GuestPhone, false); err != nil {
		return err
	}
	return validatePublicMetadata(req.Metadata)
}

func validatePublicMetadata(metadata map[string]any) error {
	if len(metadata) > 20 {
		return validationError("metadata must contain no more than 20 keys")
	}
	raw, err := json.Marshal(defaultMetadata(metadata))
	if err != nil {
		return validationError("metadata must be valid JSON")
	}
	if len(raw) > 4096 {
		return validationError("metadata must be 4 KiB or smaller")
	}
	return validateMetadataValue(metadata, 0)
}

func validateMetadataValue(value any, depth int) error {
	if depth > 3 {
		return validationError("metadata nesting is too deep")
	}
	switch typed := value.(type) {
	case nil, bool, float64, json.Number:
		return nil
	case string:
		if utf8.RuneCountInString(typed) > 500 {
			return validationError("metadata strings must be 500 characters or fewer")
		}
		return nil
	case []any:
		if len(typed) > 20 {
			return validationError("metadata lists must contain no more than 20 items")
		}
		for _, item := range typed {
			if err := validateMetadataValue(item, depth+1); err != nil {
				return err
			}
		}
		return nil
	case map[string]any:
		for key, item := range typed {
			if length := utf8.RuneCountInString(strings.TrimSpace(key)); length < 1 || length > 64 {
				return validationError("metadata keys must be between 1 and 64 characters")
			}
			if err := validateMetadataValue(item, depth+1); err != nil {
				return err
			}
		}
		return nil
	default:
		return validationError("metadata contains an unsupported value")
	}
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

// recoveryStatusTransitionAllowed keeps case state changes explicit and prevents
// terminal cases from being silently reopened.
func recoveryStatusTransitionAllowed(from, to string) bool {
	if from == to {
		return true
	}
	allowed := map[string]map[string]bool{
		"new":                        {"open": true, "assigned": true, "spam": true, "duplicate": true, "archived": true},
		"open":                       {"assigned": true, "waiting_on_guest": true, "waiting_on_internal_action": true, "offer_pending": true, "resolved": true, "closed": true, "spam": true, "duplicate": true},
		"assigned":                   {"waiting_on_guest": true, "waiting_on_internal_action": true, "offer_pending": true, "resolved": true, "closed": true, "spam": true, "duplicate": true},
		"waiting_on_guest":           {"open": true, "assigned": true, "offer_pending": true, "resolved": true, "closed": true},
		"waiting_on_internal_action": {"open": true, "assigned": true, "offer_pending": true, "resolved": true, "closed": true},
		"offer_pending":              {"open": true, "assigned": true, "waiting_on_guest": true, "resolved": true, "closed": true},
	}
	return allowed[from][to]
}

func enforceCampaignMetadata(metadata map[string]any, flyer bool, rating int) map[string]any {
	if metadata == nil {
		metadata = map[string]any{}
	}
	if flyer {
		metadata["campaign_type"] = "flyer_giveaway"
		metadata["follow_up_required"] = rating < 5
		metadata["public_review_prompt"] = rating == 5
	}
	return metadata
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
