# Slice 2

## Product feature

Printed flyer giveaway survey campaign.

This is the first product workflow that matters right now: a Heard restaurant customer creates a QR-backed survey for a printed flyer placed on takeout bags. The guest has no known identity and no known purchase context when they receive the flyer.

## Flow

1. Restaurant admin creates a campaign for a location.
2. Heard generates an opaque survey link and qurl-backed QR asset.
3. The restaurant prints the QR code and SMS keyword on flyers.
4. Guest scans the QR code or texts the keyword.
5. Guest taps one of five faces to rate the experience from 1 to 5.
6. Ratings 1 through 4 collect feedback and contact information before completion.
7. Ratings 1 through 4 create a recovery case.
8. Rating 5 prompts Google and Yelp review links first.
9. Rating 5 then collects contact information for the giveaway entry.

## Rules

- No POS integration.
- No order identity.
- No customer identity before scan/text.
- Campaign is the attribution context.
- Phone or email is required for giveaway entry.
- Marketing consent is separate from transactional follow-up.
- Public review prompting only happens after a 5 rating.
- Anything below 5 is follow-up required.

## Implemented surfaces

- Admin campaign builder: `/admin/campaigns`
- Guest campaign survey: `/f/{token}`
- Seed campaign token: `demo-heard`
- Recovery inbox: `/admin/recovery`

## qurl boundary

Heard must not render QR codes itself. It only sends the destination URL to qurl when `QURL_BASE_URL` is configured and stores the asset reference qurl returns.

When qurl is not configured, Heard still creates the campaign and survey link, but QR asset generation is shown as unavailable.
