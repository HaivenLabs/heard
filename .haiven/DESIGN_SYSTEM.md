# Haiven Design System

This document defines the shared design direction for Haiven products.

It is not a full component library yet.

It is the product experience standard every Haiven repo should follow.

---

## Product feel

Haiven products should feel:

- sharp
- calm
- modern
- useful
- trustworthy
- fast
- premium without being precious
- simple without feeling cheap

The UI should never feel like a developer demo.

---

## Core UX principles

### Mobile-first, responsive everywhere

Every user-facing surface should be designed mobile-first and scale cleanly to tablet, laptop, and desktop.

Avoid separate device-specific variants unless a platform constraint makes that unavoidable.

Responsive behavior is part of the feature, not deferred polish.

### Screenshot-ready from the first usable slice

The first usable slice should be good enough to show.

It does not need every feature.

It does need to look intentional.

### Fewer surfaces, better surfaces

Do not create separate guest, admin, test, and internal experiences that all solve the same workflow differently.

Use shared primitives and patterns.

Internal test harnesses must be clearly labeled and separated from production UX.

### Useful states

Every meaningful surface needs:

- empty state
- loading state
- error state
- populated state
- success/confirmation state where applicable

### Clear hierarchy

Every page should make the next action obvious.

Use clear hierarchy:

- page title
- short explanatory copy
- primary action
- secondary action
- status/context
- supporting details

---

## Visual direction

Default direction:

- clean spacing
- strong typography
- minimal clutter
- soft hierarchy
- thoughtful contrast
- rounded cards where appropriate
- sharp tables where appropriate
- clear form controls
- restrained motion
- polished but not over-designed

Avoid:

- default browser forms
- giant walls of text
- ambiguous icons
- hidden required fields
- unclear disabled states
- low-contrast gray soup
- admin panels that look like raw database screens

---

## Interaction standards

Forms:

- Use human-readable labels.
- Avoid asking users for raw IDs.
- Use dropdowns, pickers, search, or lookup controls when selecting known entities.
- Validate early and clearly.
- Preserve user input when validation fails.
- Show successful submission states.

Tables and lists:

- Include useful empty states.
- Include pagination where large.
- Show status clearly.
- Support search/filter where useful.
- Avoid showing irrelevant internal implementation details.

Errors:

- User-facing errors should say what happened and what to do next.
- Internal details should go to logs, not the user.
- Provider failures should degrade gracefully where possible.

---

## Accessibility baseline

Required:

- semantic HTML where possible
- accessible labels for form controls
- keyboard navigability
- visible focus states
- sufficient contrast
- reduced-motion friendliness where motion exists
- no critical information conveyed only by color

---

## Shared design tokens

Initial token categories:

- color
- typography
- spacing
- border radius
- elevation/shadow
- breakpoints
- motion
- z-index

Do not overbuild the token system before products need it.

Start with consistent primitives and extract shared packages when duplication proves the need.

---

## Product family consistency

Products should feel related, but not identical.

Passage should feel simple, secure, and frictionless.

qurl should feel creative, fast, visual, and trustworthy.

Heard should feel calm, restaurant-friendly, polished, and operationally useful.

All should feel like Haiven.
