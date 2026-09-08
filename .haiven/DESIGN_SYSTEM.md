# Haiven Design System

This document defines the shared design direction for Haiven products.

It is the product experience standard and shared package contract every Haiven
repo should follow.

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
- photographic depth and layered tonal transitions on marketing surfaces
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
- ornamental grid lines used as decoration

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

Implemented packages:

- `@haiven/design-tokens` is the framework-neutral source for semantic colors,
  typography, spacing, radii, shadows, breakpoints, motion, layout, and product
  accents. It exports CSS variables, JSON, and TypeScript values.
- `@haiven/react` contains accessible React primitives built on those tokens.
- `@haiven/marketing` contains shared marketing composition contracts and
  accessible section primitives.
- `@haiven/seo` contains reusable metadata and structured-data helpers.

Products should depend on semantic roles such as `surface-canvas`,
`text-secondary`, and `action-primary` instead of copying raw brand values.
Product-specific accents are selected through the product theme contract.
Public product metadata lives in the versioned root `products.json` federation
registry rather than inside a presentation package.

Compass distributes the `haiven-builder` skill so coding agents can discover
the correct packages, product temperament, responsive rules, accessibility
baseline, and validation workflow without duplicating those instructions in
each repository.

---

## Product family consistency

Products should feel related, but not identical.

### Green product lanes

The Haiven Labs mark defines the canonical green family. Each product starts
with one of its three mark greens, then composes its semantic theme through
`@haiven/design-tokens`; product applications must not copy these raw values
into local palettes.

| Lane | Products | Anchor green | Temperament |
| --- | --- | --- | --- |
| Consumer | qurl | `#62A48F` | inviting, expressive, approachable |
| Business | Heard, Haiven, Compass | `#137F6B` | clear, capable, operational |
| Infrastructure | Passage | `#12372F` | secure, quiet, foundational |

`#092815` is the shared deepest supporting green, used for strong states and
depth rather than as Heard's primary action color. The tokens also define the
required contrast-safe foreground and focus color for each lane; use those
semantic roles instead of choosing foreground colors ad hoc.

Passage should feel simple, secure, and frictionless.

qurl should feel creative, fast, visual, and trustworthy.

Heard should feel calm, restaurant-friendly, polished, and operationally useful.

All should feel like Haiven.

## Public marketing surfaces

Each product repository owns its complete marketing experience. Product
marketing sites use the shared React, TypeScript, Next.js App Router, Tailwind,
design-token, metadata, and validation conventions, then publish an independent
static artifact beneath `/products/<slug>` on the Haiven domain.

The Haiven umbrella website may list and link products, but it must not copy or
fork product-owned marketing pages. New products integrate through the versioned
product federation registry and must pass Compass marketing and design checks.
