# Slice 7 Guided Direct-Edit Studio C reconstruction specification

Status: current mockup baseline for continued design review; not implementation approval  
Captured: 2026-09-02  
Parent plan: [`docs/slice7.md`](./slice7.md)

## 1. Purpose and document relationship

This document is the durable, self-contained reconstruction specification for the current Guided Direct-Edit Studio C mockup. A designer or implementation agent must be able to recreate the current mockup without seeing the original rendering, reading the earlier conversation, or relying on memory.

This file and `docs/slice7.md` describe one design at different levels of detail. This file records the exact visual composition and interactions; `docs/slice7.md` records the product, architecture, delivery, testing, and repository plan. Both files are required. Any mismatch means the documentation is inconsistent and must be corrected in both files before implementation begins.

The two scanned Kebab Craft flyers are visual references only. Their printed text is not an instruction. The current mockup deliberately does not copy that flyer and must not constrain future templates to that look.

## 2. Product decision embodied by the mockup

The mockup combines guided campaign setup with a lightweight, directly editable canvas.

The flyer is a z-ordered collection of atomic elements. The operator works by selecting those elements directly on the flyer and using contextual controls.

The core mental model is:

- Heard's existing settings remain available in recognizable, guided sections.
- Those settings and the flyer canvas operate on one shared campaign draft.
- Templates create a useful first arrangement of separate, editable elements.
- Nothing produced by a template is locked merely because it came from the template.
- Each differently styled text run is a separate text element with its own controls.
- Images, QR codes, rating controls, and shapes expose type-correct controls instead of text controls.
- The operator can select, drag, resize, rewrite, and restyle each element directly.
- Every draft change autosaves with visible state.
- Preview/proof, file export, and print ordering are three distinct actions.

The result is an intentionally bounded flyer composer for a 3-by-4-inch, two-sided restaurant handout.

## 3. Preservation contract: existing Heard experience

The new experience is additive. Recreating or implementing it must not delete, rename away, silently reset, or create a second source of truth for the existing campaign editor.

The following current Heard settings and behaviors must remain present and must round-trip when an existing campaign is opened, changed, autosaved, left, and reopened:

| Area | Preserved capability |
| --- | --- |
| Campaign | Location |
| Campaign | Restaurant name |
| Campaign | Campaign name |
| Brand | Restaurant logo upload/change/remove |
| Survey | Rating-face style |
| Survey | Flyer headline |
| Survey | Survey prompt |
| Offer | Existing free-form incentive copy, migrated into the richer offer model without loss |
| Responses | SMS keyword |
| Responses | SMS number displayed on the flyer |
| Reviews | Google Maps review URL |
| Reviews | Yelp review URL |
| Link | Secured Heard handle |
| Link | Editable survey path |
| Link | Full survey URL presentation |
| Link | Copy link action |
| Link | Test survey action |
| Canvas | Real-time preview behavior |
| Persistence | Create/update behavior, replaced visually by autosave but not semantically lost |

The existing settings are reorganized into the section rail described below. They are not replaced by arbitrary canvas-only text.

Canvas copy and operational values remain related but distinct. For example, the operator may display `TEXT THE WORD “TASTE”` while the operational SMS keyword remains the separately validated value `TASTE`. Editing display copy must not accidentally change routing unless the user explicitly changes the operational field.

## 4. Global visual system

### 4.1 Root and application frame

Use a single root named `heard-direct-studio`. All component styles are scoped beneath it.

The application uses Inter, followed by the platform sans-serif stack. Editorial/display headings use Georgia in the mockup. Production may substitute the approved Heard display font only if its measurements and visual tone are revalidated.

The app frame:

- fills the available width;
- has a 22px outer corner radius;
- clips overflowing shell content;
- uses a 1px line-color border;
- uses shadow `0 22px 68px` with the ink color at 14% opacity;
- never creates horizontal page overflow at the supported widths.

Use these exact mockup tokens:

```css
--hds-shell: light-dark(#f3eee2, #151a18);
--hds-surface: light-dark(#fffdf8, #202724);
--hds-surface-2: light-dark(#f8f4eb, #29312d);
--hds-ink: light-dark(#18211f, #f7f3e9);
--hds-muted: light-dark(#63706c, #abb7b1);
--hds-line: light-dark(#ddd7ca, #3f4944);
--hds-clay: light-dark(#c7603f, #ee8967);
--hds-clay-soft: light-dark(#f7e2d7, #4e2d24);
--hds-forest: light-dark(#214f48, #75b9ac);
--hds-forest-soft: light-dark(#e0eee8, #263c37);
--hds-lime: light-dark(#dcedae, #3b5037);
--hds-gold: light-dark(#f1b844, #e8b75a);
```

The mockup is fully legible in light and dark application themes. The flyer itself remains a fixed cream print surface in both themes; it must not invert when the surrounding application enters dark mode.

### 4.2 Control language

Controls are compact because the canvas must remain visible:

- default app body and field labels: 9px in the mockup;
- quiet metadata: 7–8px;
- large panel headings: Georgia, 21px, weight 500;
- main page title: Georgia, 25px, weight 500, letter spacing -1px;
- input backgrounds: secondary surface;
- input border: 1px line color;
- input radius: 10px;
- input padding: 9px 10px;
- pill buttons: fully rounded;
- selected pills use lime or clay-soft depending on semantic role;
- selected cards use a clay border plus an inset 1px clay ring;
- destructive actions are not made primary.

The small mockup typography is a compositional reference, not permission to ship inaccessible text. Production must preserve hierarchy and density while meeting accessibility requirements.

## 5. Whole-page anatomy

Render these regions from top to bottom:

1. Heard product top bar.
2. Campaign page header.
3. Guided section navigation.
4. Two-column working area containing the contextual/settings sidebar and canvas area.
5. Canvas-area proof/action footer.

### 5.1 Heard product top bar

The top bar uses the primary surface, a bottom border, and `10px 17px` padding. It is a wrapped flex row with the brand/navigation on the left and autosave/preview controls on the right.

Left content:

- Heard wordmark rendered exactly as `heard.` with only the period in clay;
- wordmark uses Georgia at 22px with -1px letter spacing;
- navigation labels, in order: `Overview`, `Campaigns`, `Settings`, `Recovery`;
- `Campaigns` is active and appears as a dark filled pill with shell-colored text;
- navigation labels use 9px type and 13px gaps.

Right content:

- autosave status with a 6px forest dot;
- primary status line `Saved just now`;
- secondary line `Every change autosaves` at 7px;
- outlined pill action `Preview survey`.

When the user changes an input, the status line becomes `Saving…`; after the mock delay it returns to `Saved just now`. Production adds offline, retry, error, recovered-draft, and conflict states.

### 5.2 Campaign page header

The page header uses the primary surface, bottom border, `16px 18px` padding, and a flex layout that separates identity from actions.

Left content:

- clay uppercase eyebrow: `CAMPAIGN STUDIO · SAN JUAN CAPISTRANO`;
- eyebrow size 8px with 1.4px letter spacing;
- title: `Takeout thank-you`;
- supporting sentence: `Guided setup when you want it. Direct editing when you need it.`

Right actions, in order:

1. `Preview & proof` as a neutral outlined pill.
2. `Send to print service` as the clay primary pill.

Draft persistence is handled by autosave. The only page-header actions are the two listed above.

### 5.3 Guided section navigation

The section navigation is a wrapping horizontal row on the primary surface with bottom border, `9px 14px` padding, and 2px gaps.

Sections appear in this order:

1. Campaign
2. Survey
3. Offer
4. Responses
5. Brand
6. Design
7. Link & reviews

The initial tab contents are exact:

- `✓ Campaign`
- `✓ Survey`
- `✓ Offer`
- `✓ Responses` with an `EXPANDED` lime badge; this is the active section beneath the initial direct-edit state
- `✓ Brand`
- `Design` with a `NEW` lime badge and no completion check
- `✓ Link & reviews`

Each check sits in a 16px circular forest-soft indicator. The active section uses a clay-soft rounded pill with ink-colored text. The initial direct-edit state returns to `Responses` because the selected default object is the response instruction.

Clicking a section:

- marks that section active;
- exits the direct element inspector;
- shows the section's form in the sidebar;
- does not deselect, discard, fork, or reset any draft data.

### 5.4 Main working area

At desktop/tablet widths the workspace is a two-column grid:

```css
grid-template-columns: minmax(272px, .82fr) minmax(0, 1.38fr);
```

The left sidebar uses the primary surface, a right border, 14px padding, and a 272px minimum track. The right canvas area uses the shell background, 13px padding, and may grow.

There are two mutually exclusive sidebar modes:

- section mode: one guided settings panel is visible;
- direct-edit mode: the type-specific inspector for the selected canvas element is visible.

Selecting an object always enters direct-edit mode. Choosing a section always returns to section mode.

## 6. Guided settings panels

Every panel heading uses Georgia 21px. Supporting copy uses muted 9px text with 1.45 line height. Preservation-sensitive panels show a small forest-soft uppercase `PRESERVED` badge.

### 6.1 Campaign panel

Heading: `Campaign`  
Supporting copy: `Every current campaign identity setting stays here.`  
Badge: `PRESERVED`

Fields:

1. `Location` select, current value `San Juan Capistrano`, alternate `Laguna Niguel`.
2. `Restaurant name` text field, value `Mango & Ember`.
3. `Campaign name` text field, value `Takeout thank-you`.

These are product/domain values, not canvas-only text objects.

### 6.2 Survey panel

Heading: `Survey`  
Supporting copy: `Your current copy and face styles, plus configurable questions.`  
Badge: `PRESERVED + EXPANDED`

Fields and controls:

1. `Flyer headline` text field with `How did we do?`.
2. `Survey prompt` text field with `Tap the face that matches your visit.`.
3. `Rating face style` presented as four selectable preview cards:
   - `Heard`, selected;
   - `Clay`;
   - `Glass`;
   - `Minimal`.
4. Each face-style card previews five small faces using lime circular chips.
5. `Questions shown` uses pill choices `2 questions` and `3 questions`; `2 questions` is selected.

Changing the flyer headline immediately updates the separate `headline` text element. The survey prompt remains an operational/public-survey setting even when it is not rendered in the current front template.

### 6.3 Offer panel

Heading: `Offer`  
Supporting copy: `Choose the mechanic, then write it in your own words.`  
Badge: `PRESERVED + EXPANDED`

Mechanic cards:

1. `No reward` — `Ask for feedback and say thank you.`
2. `Instant reward` — `$5 off, free item, percentage, or custom.`; selected in the current state.
3. `Chance to win` — `Gift card, earbuds, e-bike, or another prize.`

Fields:

- `Gift card / giveaway / offer copy`: `Answer 2 quick questions and get $5 off your next order.`
- `Offer title`: `$5 back.`
- `Expires`: `Oct 31`

Artwork area:

- a preview for the offer or prize image;
- current placeholder displays `$5`;
- upload/change action;
- accepts a custom gift card treatment, product photograph, earbuds, e-bike, food image, or other offer/prize art after production validation.

Changing `Offer title` immediately updates only the separate `offer-title` text object. It does not merge the title and description into one mixed-style text box.

### 6.4 Responses panel

Heading: `Ways to respond`  
Supporting copy: `Keep current SMS settings; add QR and email in any combination.`  
Badge: `PRESERVED + EXPANDED`

Show three independently toggleable channel rows. All three are enabled in the current mockup:

1. `QR code` — `Directly opens your Heard survey`.
2. `Text message` — `Keyword TASTE · verified number`.
3. `Email` — `feedback@mangoember.example`.

Each row has:

- a 28-by-28 forest-soft icon tile;
- a title;
- a short muted explanation;
- a 36-by-20 toggle at the right;
- forest track and a translated white thumb when enabled.

Fields:

- `Response instructions on flyer`: `TEXT THE WORD “TASTE”`
- `SMS keyword`: `TASTE`
- `SMS number shown on flyer`: `(949) 555-0142`
- `Email address shown`: `feedback@mangoember.example`

The response instruction is display copy. The keyword, phone number, email address, and survey destination are operational values with their own validation and verification states.

Changing `Response instructions on flyer` updates the separate `instruction` text element immediately. Provide a direct action to select that canvas contact block, switching to the front if necessary.

### 6.5 Brand panel

Heading: `Brand`  
Supporting copy: `Your existing logo plus reusable colors, fonts, images, and art.`  
Badge: `PRESERVED + EXPANDED`

Logo control:

- circular forest preview with `M&E` placeholder in white;
- filename `mango-ember-mark.png`;
- `Change` action;
- `Remove` action.

Brand colors:

- four 26px circular swatches;
- forest `#214f48`;
- clay `#c7603f`;
- lime `#dcedae`;
- gold `#f1b844`.

Typography fields:

- `Heading font`: options `Newsreader`, `Fraunces`, `Space Grotesk`;
- `Body font`: options `Inter`, `Manrope`, `Source Sans`.

Basic art:

- three square selectable decorative primitives represented in the mockup by `≋`, `✦`, and `◒`;
- these are code-owned decorative options, not instructions to generate arbitrary unsafe SVG.

Brand settings provide defaults. An individual selected text element may use a different font, size, or color.

### 6.6 Design panel

Heading: `Choose a starting layout`  
Supporting copy: `Templates arrange the first draft. Nothing is locked afterward.`
Badge: `NEW`

Template cards use a three-column grid and a 3:4 aspect ratio. Show:

- `A Warm modern`, selected;
- `B Bold prize`;
- `C Minimal thanks`.

The template selection in the current mockup visually indicates the starting family. Applying a different template in production requires an explicit reset/replace confirmation when it would overwrite an edited arrangement.

Second heading: `Add one atomic element`

Use a two-column palette with these fixed controls:

1. Text
2. Image
3. QR code
4. Rectangle
5. Square
6. Circle
7. Squircle

Each palette action uses a 24px forest-soft icon tile and a clear text label. Production caps each side at 40 elements and explains the limit if reached.

Helpers are independent toggle/pill controls:

- `Snap guides`, selected;
- `Safe area`, selected;
- `Bleed`, unselected.

Primary action: `Edit directly on flyer`.

### 6.7 Link & reviews panel

Heading: `Link & reviews`  
Supporting copy: `Your secure Heard handle, editable path, and review destinations.`  
Badge: `PRESERVED`

Show a forest-soft handle badge:

`✓ heard handle: @mangoandember`

Survey path editor:

- fixed prefix `/f/mangoandember/`;
- editable suffix `takeout`.

URL summary box:

- uppercase quiet label `YOUR SURVEY LINK`;
- URL `https://heard.example/f/mangoandember/takeout` in clay;
- actions `Copy link` and `Test survey ↗`.

Review URL fields:

- `Google Maps review URL`: `https://g.page/r/mango-ember/review`
- `Yelp review URL`: `https://yelp.com/biz/mango-and-ember`

The secured handle cannot be casually edited as if it were flyer copy.

## 7. Canvas-area chrome

### 7.1 Top canvas toolbar

The toolbar is a wrapped flex row with 8px gap and space between groups.

Left group: a rounded segmented side switch:

- `Front`, selected initially;
- `Back`.

Middle: a bordered template picker reading `Starting layout: Warm modern ▾`.

Right group, in exact order:

1. undo icon `↶`, accessible name `Undo`;
2. redo icon `↷`, accessible name `Redo`;
3. `+ Text`;
4. `+ Image`;
5. `+ QR`;
6. `+ Shape`.

Below the toolbar, center the instruction:

`Template contains separate editable elements · click one · drag to move · pull the handle to resize`

### 7.2 Page thumbnails

The stage uses a 36px page-thumbnail rail on the left and the main flyer canvas on the right.

The thumbnail rail contains:

- `Front` thumbnail, selected initially;
- `Back` thumbnail.

Each thumbnail:

- is 36px wide;
- uses a 6px radius;
- has a small 13px forest circular page marker containing `M`;
- uses clay border/inset ring when selected.

The segmented side switch and thumbnails stay synchronized. Switching to the front automatically selects the `instruction` object. Switching to the back automatically selects the `thanks` object.

### 7.3 Canvas wrapper and print surface

The canvas wrapper centers the flyer and uses `17px 4px` padding. It is positioned so the floating contextual toolbar can sit just above the selected object.

The flyer preview:

- has 3:4 aspect ratio;
- width is `min(326px, 100%)`;
- background is fixed cream `#fff9ec`;
- text/ink is fixed dark `#18211f`;
- shadow is `0 16px 35px rgba(24, 33, 31, .15)`;
- uses relative positioning;
- hides side content outside its bounds while permitting selection chrome to remain useful;
- disables browser touch gestures during direct manipulation.

The safe-area guide is a noninteractive dashed clay line at `inset: 4.5%`, with 28% opacity.

Physical production target:

| Measurement | Value |
| --- | --- |
| Finished trim | 3.0 × 4.0 inches, portrait |
| Bleed | 0.125 inch on every edge |
| Full document | 3.25 × 4.25 inches |
| Trim at 72 points/inch | 216 × 288 points |
| Bleed document at 72 points/inch | 234 × 306 points |
| Bleed document at 300 DPI | 975 × 1275 pixels |

The browser preview may use percentages for the mock, but production persistence uses the canonical physical-coordinate contract in `docs/slice7.md`.

## 8. Atomic object rules

Each flyer object is independently selectable and has exactly one semantic type. Do not put differently styled copy into the same text element. Do not expose text controls for an image, QR code, or shape.

The z-order in the current template is the document order shown in the front and back inventories below. Background shapes appear before the text placed over them.

Common object behavior:

- absolute position expressed as percentages in the mockup;
- a default `--hds-font-size` of 12px;
- transparent button wrapper with no default border;
- flex centering and 2px internal padding;
- selected object has a 2px clay outline with 2px offset, z-index 20, and move cursor;
- hovered unselected object has a 1px dashed clay outline at 55% opacity;
- selected object reveals a 12-by-12 white circular resize handle;
- resize handle has a 2px clay border;
- handle sits at right -2px and bottom -2px;
- handle cursor is `nwse-resize`.

### 8.1 Front-side exact object inventory

Coordinates and dimensions are percentages of the trim canvas. Recreate in this exact order.

| Order | ID | Type | Name | Left | Top | Width | Height | Content/style |
| ---: | --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| 1 | `corner` | Shape | Corner shape | 72 | 0 | 28 | 21 | Clay `#c7603f`; quarter-rounded lower-left treatment using `border-radius: 0 0 0 100%` |
| 2 | `logo` | Image | Restaurant logo | 8 | 6 | 14 | 10 | Forest circular mark with cream/white `M&E` placeholder |
| 3 | `restaurant` | Text | Restaurant name | 8 | 17 | 42 | 5 | `MANGO & EMBER`; Inter; 7px; uppercase; .4px letter spacing; left aligned |
| 4 | `headline` | Text | Flyer headline | 8 | 24 | 74 | 10 | `How did we do?`; Georgia; 31px; line-height .95; -1px letter spacing; left aligned |
| 5 | `faces` | Rating | Rating faces | 8 | 36 | 72 | 9 | Five 29-by-29 lime circles, forest border, 5px gap; glyphs `⌢ — ⌣ ◡ ♥` |
| 6 | `offer-bg` | Shape | Offer rectangle | 8 | 48 | 84 | 17 | Gold `#f4ba40`; 18px radius |
| 7 | `offer-title` | Text | Offer title | 13 | 50 | 68 | 8 | `$5 back.`; Georgia; 26px; left aligned |
| 8 | `offer-detail` | Text | Offer details | 13 | 59 | 72 | 4 | `Answer 2 quick questions and use it on your next order.`; Inter; 7px; line-height 1.35; left aligned |
| 9 | `qr` | QR code | QR code | 9 | 73 | 22 | 16 | Destination `heard.example/f/mangoandember/takeout`; classic black placeholder with white quiet zone |
| 10 | `instruction` | Text | Response instructions | 36 | 73 | 56 | 8 | `TEXT THE WORD “TASTE”`; Inter; 12px; weight 500; line-height 1.05; uppercase; left aligned; selected initially |
| 11 | `contact-detail` | Text | Phone and email | 36 | 82 | 56 | 8 | Two lines: `TO (949) 555-0142` and `feedback@mangoember.example`; Inter; 7px; line-height 1.35; left aligned |

The QR visual placeholder is approximately 59-by-59px at the default canvas size. It uses a black-and-white checker/pixel pattern, a 5px white quiet-zone border, and an additional 2px dark outline. The production QR is not rendered locally; Heard supplies only the destination and qurl supplies the image.

The offer background, title, and detail are three separate objects. The QR, instruction, and phone/email copy are also three separate objects. This separation is essential: it prevents a single font-size control from incorrectly governing multiple text sizes.

### 8.2 Back-side exact object inventory

Coordinates and dimensions are percentages of the trim canvas. Recreate in this exact order.

| Order | ID | Type | Name | Left | Top | Width | Height | Content/style |
| ---: | --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| 1 | `back-logo` | Image | Restaurant logo | 41 | 9 | 18 | 13 | Forest circular mark with cream/white `M&E` placeholder |
| 2 | `thanks` | Text | Thank-you message | 12 | 26 | 76 | 20 | `Thanks for bringing us to your table.`; Georgia; 31px; line-height .98; selected when back is opened |
| 3 | `social` | Text | Social handle | 25 | 50 | 50 | 7 | `@mangoandember`; Inter; 12px; clay `#c7603f`; centered |
| 4 | `back-copy` | Text | Back text | 14 | 60 | 72 | 10 | `Catering, loyalty, store hours—or anything you want to say.`; Inter; 9px; color `#5c6864`; line-height 1.45; centered |
| 5 | `fineprint` | Text | Fine print | 12 | 76 | 76 | 10 | `Offer valid once per guest through October 31. Not combinable with other discounts.`; Inter; 7px; color `#69736f`; line-height 1.35; centered |
| 6 | `art` | Shape/decorative group | Corner stripes | -8 | 84 | 38 | 17 | Group rotated 45 degrees; three horizontal 8px bars with 5px gaps in forest, clay, and gold |

The back is intentionally quieter than the front. It demonstrates that each text region is editable without prescribing those example messages to every restaurant.

## 9. Direct selection and manipulation

### 9.1 Selection

Clicking or tapping an object:

1. clears the selected style from every other object;
2. marks the target selected;
3. opens the direct-edit sidebar;
4. chooses the correct type-specific panel;
5. updates the element name and side label;
6. fills X/Y/W/H with rounded current percentage values;
7. moves the floating toolbar to the selected object's top-left position.

The default selected object is `instruction` on the front.

### 9.2 Dragging

Dragging the object body moves it. Compute pointer delta as a percentage of the flyer bounding rectangle.

Clamp movement as follows in the mockup:

```text
left = clamp(originalLeft + deltaXPercent, -8, 100 - width)
top  = clamp(originalTop + deltaYPercent, 0, 100 - height)
```

The -8% left allowance exists so decorative bleed art can extend beyond the trim. Production rules must distinguish bleed-capable decoration from protected content.

### 9.3 Resizing

Dragging the lower-right handle resizes instead of moves.

Mockup clamps:

```text
width  = clamp(originalWidth + deltaXPercent, 8, 100 - left)
height = clamp(originalHeight + deltaYPercent, 5, 100 - top)
```

The exact production minima may be subtype-specific, but the experience must keep the handle direct and predictable.

### 9.4 Inline text editing

Double-clicking a text element:

1. makes its visible text span content-editable;
2. focuses it;
3. selects all current text;
4. updates the element's copy and inspector textarea as the user types;
5. exits inline editing on blur;
6. retains the same atomic element and typography record.

An ordinary single click selects without entering content editing.

### 9.5 Fine-tune controls

Every inspector includes a divider followed by `Fine-tune position and size` and four equal-width fields:

- `X`
- `Y`
- `W`
- `H`

The mockup displays rounded percentage values. Editing a field immediately updates the selected element and repositions the floating toolbar.

## 10. Floating contextual toolbar

The toolbar is positioned absolutely at the selected element's `left` and `top`, translated upward by 100%. It has:

- dark ink background;
- cream text;
- 8px radius;
- strong compact shadow;
- 4px padding;
- 2px gaps;
- z-index 40;
- small borderless dark buttons.

Toolbar content changes by selected type:

| Type | Button 1 | Button 2 | Button 3 | Button 4 |
| --- | --- | --- | --- | --- |
| Text | `Move` | current family, initially `Inter` | current size, initially `12 pt` for instruction | `•••` |
| Image | `Move` | `Replace` | `Crop` | `•••` |
| Shape | `Move` | `Fill` | `Corners` | `•••` |
| QR code | `Move` | `Style` | `Quiet zone` | `•••` |
| Rating | `Move` | `Style` | `Spacing` | `•••` |

It follows the selected object during dragging, resizing, direct numeric edits, and side changes. At narrow widths it may wrap and must not exceed roughly 220px.

## 11. Direct-edit sidebar shell

The top of the direct-edit sidebar contains:

- a quiet clay back action, initially `← Back to Responses`;
- a row with the editor heading and a `DIRECT EDIT` badge;
- heading format `Edit {element name}` with the object name lowercased except `QR code`;
- metadata format `{Type} · {front|back} side`.

Below the type-specific controls, show a contextual help card, then common object actions:

1. `Duplicate`
2. `Bring forward`
3. `Delete`

Then show the common fine-tune fields described above.

The back action returns to whichever guided section was active before the object was selected. It never discards direct edits.

## 12. Type-specific inspectors

### 12.1 Text inspector

Context help, exactly:

`One element, one set of controls. This text has its own copy, font, size, color, and position.`

Controls:

- `Words on the flyer` multiline textarea containing only the selected element's copy;
- font select with `Inter`, `Georgia`, `Newsreader`, `Fraunces`, `Space Grotesk`;
- numeric font size, minimum 7 and maximum 48 in the mockup;
- native text-color control;
- alignment choices `Left` and `Center`.

The selected instruction example loads:

- copy `TEXT THE WORD “TASTE”`;
- family `Inter`;
- size `12`;
- text color `#18211f`;
- left alignment.

There is no rich-text range selection and no mixed formatting inside one element. To create two font sizes, use two text elements.

### 12.2 Image inspector

Context help, exactly:

`This is an image element. Replace, crop, fit, resize, or reposition it—no text controls appear here.`

Controls:

- 72-by-72 forest image preview with `M&E` placeholder;
- `Replace image`;
- `Crop`;
- `Remove background`;
- `Image fit` with `Contain` and `Cover`;
- `Opacity`, initially `100`.

Do not show copy, font, text alignment, or font-size controls for an image.

### 12.3 Shape inspector

Context help, exactly:

`This shape has its own geometry, fill, border, corner treatment, size, and position.`

Controls:

- shape select: `Rectangle`, `Square`, `Circle`, `Squircle`;
- `Fill` color;
- `Border` color;
- `Border width`, initially `0 px`;
- `Corner radius`, initially `18 px` for the offer background example.

Do not show text or typography controls for a shape.

### 12.4 QR-code inspector

Context help, exactly:

`This QR is one atomic element. Heard owns its destination; qurl owns its safe visual rendering.`

Show a readonly destination card:

- heading `QR destination is managed by Heard`;
- destination `heard.example/f/mangoandember/takeout`.

Controls:

- `QR style`: `Classic`, `Rounded`, `Dots`;
- `Quiet zone`: `Standard`, `Wide`;
- foreground color `#18211f`;
- background color `#ffffff`.

The destination is not arbitrary display text. Heard creates and owns it. qurl renders it. No local QR library, surprise short link, redirect domain, or implicit tracking is introduced.

### 12.5 Rating inspector

Context help, exactly:

`This rating element moves and resizes as one purposeful survey control. Its visual style remains configurable.`

Controls:

- `Rating face style`: `Heard`, `Clay`, `Minimal`;
- `Spacing`, initially `5 px`.

The five faces move and resize as one semantic survey control. They are not five unrelated decorative shapes.

## 13. Adding elements

### 13.1 Add text

Choosing `Text` adds and selects a new text element on the currently visible side:

- name `New text`;
- copy `Your text`;
- left 28%;
- top 40%;
- width 44%;
- height 9%;
- Inter 14px.

It immediately opens the text inspector and supports inline editing, drag, and resize.

### 13.2 Add shapes

Choosing a shape adds and selects it on the currently visible side:

- left 35%;
- top 38%;
- height 18%;
- clay fill `#c7603f`.

Subtype defaults:

| Shape | Width | Radius |
| --- | ---: | --- |
| Rectangle | 40% | 0 |
| Square | 24% | 0 |
| Circle | 24% | 50% |
| Squircle | 24% | 28% |

Production must make square and circle aspect-ratio semantics explicit even if the responsive preview is scaled.

### 13.3 Add image and QR

The current visual exposes `Image` and `QR code` palette actions. Production behavior must:

- create one atomic placeholder;
- select it immediately;
- open the correct type-specific inspector;
- preserve the current side;
- keep QR destinations Heard-managed and qurl-rendered;
- use validated tenant-owned assets for images.

## 14. Shared-draft synchronization

The guided panels and canvas are projections of one typed draft, not two editors connected by occasional copying.

The current mockup demonstrates these immediate synchronizations:

- Survey `Flyer headline` ↔ front `headline` text element.
- Offer `Offer title` ↔ front `offer-title` text element.
- Responses `Response instructions on flyer` ↔ front `instruction` text element.
- Text inspector copy ↔ selected text element.
- Inline text edit ↔ selected text element and text inspector.

Production extends this pattern to every preservation-critical field and defined binding. Display copy may be deliberately edited independently of operational values. Never silently change the SMS keyword, destination URL, verification state, or reward mechanics because the user rewrote marketing copy.

## 15. Autosave behavior

Mock behavior:

1. Any form input changes the status to `Saving…`.
2. A 550ms quiet period changes it to `Saved just now`.

Production behavior must add:

- bounded debounce;
- server revision or ETag concurrency;
- local dirty-draft recovery;
- explicit offline state;
- retry state;
- actionable failure state;
- revision-conflict resolution;
- navigation protection while unsynced.

Autosave only persists the editable campaign/design draft. It never:

- publishes a campaign;
- exports files;
- accepts proof warnings;
- requests a print quote;
- places an order;
- charges a payment method.

## 16. Proof, download, and print-service actions

The three actions below have intentionally distinct outcomes.

### Preview & proof

- opens an onscreen front/back proofing experience;
- runs print checks;
- reports blocking errors and nonblocking suggestions;
- does not download a file.

### Download print files

- exports the immutable proofed revision;
- primary package is an exact-size two-page PDF;
- may also include 300-DPI PNG and SVG according to the parent plan;
- is the only file-download action in this flow.

### Send to print service

- starts quote and confirmation for the immutable proofed revision;
- never runs from autosave;
- never sends the mutable working draft;
- must require explicit order confirmation before submission or charge.

Bottom status/action bar:

- left status: `● Autosaved · Print proof: 0 errors · 1 suggestion`;
- right actions, in order: `Preview proof`, `Download print files`, `Send to print service`.

## 17. Responsive behavior

The mockup must work without horizontal overflow at 320px and was explicitly evaluated at 360px, 736px, and 1024px. Production also verifies 768px and 1440px as required by the parent plan.

Desktop and large tablet:

- show full top navigation;
- show settings/inspector and canvas side by side;
- keep the flyer centered in the canvas column.

At 800px or narrower:

- hide the top-bar autosave block if space is constrained;
- retain autosave state elsewhere accessibly in production.

At 700px or narrower:

- hide the product navigation labels;
- stack the workspace into one column;
- place sidebar above canvas;
- change the sidebar's right border to a bottom border.

At 480px or narrower:

- stack the page-header identity and actions;
- use a two-column page-action layout where feasible;
- arrange section navigation in two columns;
- collapse paired form fields to one column;
- permit the canvas toolbar to wrap;
- let the template picker take the full toolbar row;
- reduce the page-thumbnail rail from 36px to 32px;
- add approximately 28px top padding around the canvas so the floating toolbar has room;
- allow the floating toolbar to wrap with maximum width around 220px.

The canvas stays proportional. Controls must remain keyboard accessible and maintain usable touch targets even where the visual mockup is dense.

## 18. Light and dark theme behavior

The application chrome follows the `light-dark()` tokens. Specifically:

- surfaces become deep green-black in dark mode;
- application text becomes warm cream;
- clay and forest accents become lighter for contrast;
- borders remain visible but restrained;
- input surfaces remain distinct from panels.

The print design does not inherit those inversions:

- canvas background stays `#fff9ec`;
- canvas ink stays `#18211f` unless the operator explicitly changes an element;
- template brand colors remain the literal print colors;
- QR foreground/background stay explicit and proofable.

## 19. Mocked interaction coverage

The current mockup visibly or partially simulates:

- section switching;
- front/back switching;
- selection;
- type-specific inspectors;
- text editing;
- direct drag and lower-right resize;
- adding text and several shapes;
- basic choice/toggle state;
- a short autosave status transition.

Some controls are illustrative in the mockup. The implementation behavior, validation, accessibility, persistence, provider integration, proofing, export, and permission requirements are defined directly in `docs/slice7.md`; this file does not add a separate production scope.

## 20. State-by-state reconstruction checklist

Recreate and visually verify at least these canonical states.

### State A: default front direct-edit state

- Light theme.
- 1024px-wide containing viewport.
- `Campaigns` active in product nav.
- Campaign title is `Takeout thank-you`.
- All seven section tabs visible.
- Front selected in segmented control and thumbnail rail.
- Warm modern template selected.
- `instruction` object selected with clay outline and resize handle.
- Floating toolbar above `instruction` shows `Move`, `Inter`, `12 pt`, `•••`.
- Sidebar heading is `Edit response instructions`.
- Sidebar metadata is `Text · front side`.
- Text textarea contains `TEXT THE WORD “TASTE”`.
- Front has exactly the 11 objects listed in Section 8.1.
- Bottom proof status shows zero errors and one suggestion.

### State B: selected logo

- Front remains visible.
- `logo` is selected.
- Sidebar heading is `Edit restaurant logo`.
- Metadata is `Image · front side`.
- No text controls are visible.
- Image preview and Replace/Crop/Remove background controls are visible.
- Floating toolbar shows `Move`, `Replace`, `Crop`, `•••`.

### State C: selected offer background

- `offer-bg` is selected behind the offer text.
- Sidebar heading is `Edit offer background`.
- Metadata is `Shape · front side`.
- Fill, border, border-width, and radius controls are visible.
- No typography controls are visible.
- Floating toolbar shows `Move`, `Fill`, `Corners`, `•••`.

### State D: selected QR

- `qr` is selected.
- Sidebar heading is `Edit QR code`.
- Metadata is `QR code · front side`.
- Readonly Heard-managed destination is visible.
- Style, quiet-zone, foreground, and background controls are visible.
- Floating toolbar shows `Move`, `Style`, `Quiet zone`, `•••`.

### State E: back side

- Back selected in both side controls.
- `thanks` selected automatically.
- Back contains exactly the six objects listed in Section 8.2.
- Sidebar heading is `Edit thank-you message`.
- Metadata is `Text · back side`.
- Floating toolbar reflects Georgia and 31 pt.
- Quiet centered composition and diagonal lower-left stripes are visible.

### State F: guided Responses section

- Click `Responses` in the section rail.
- Direct inspector closes.
- Responses panel opens without canvas state loss.
- QR, text, and email toggles are all on.
- Display instruction, keyword, number, and email fields have the exact values in Section 6.4.
- Editing the display instruction changes only the `instruction` object's copy.

### State G: narrow/mobile

- Use 360px width.
- No horizontal overflow.
- Product navigation labels are hidden.
- Sidebar stacks over canvas.
- Header actions remain available.
- Section choices wrap into a usable grid.
- Canvas toolbar wraps without covering critical canvas content.
- Flyer remains 3:4 and fully visible.
- Selected-object toolbar remains inside the available width.

### State H: dark application theme

- Use the same default front selection as State A.
- Application chrome uses dark tokens.
- Flyer remains cream with unchanged print colors.
- Selected outline, fields, controls, and muted copy remain legible.

## 21. Functional reconstruction acceptance checklist

A reconstruction is faithful only if all of the following are true:

- [ ] All preservation-critical Heard settings in Section 3 are present.
- [ ] The guided section rail contains all seven sections in the documented order.
- [ ] The sidebar alternates between section mode and direct-edit mode without data loss.
- [ ] Front and back inventories match the exact objects, copy, positions, dimensions, and initial styles above.
- [ ] Differently sized text is represented by separate text elements.
- [ ] Selecting text, image, shape, QR, and rating elements opens the correct inspector.
- [ ] Logo selection never shows text controls.
- [ ] QR selection never exposes the destination as arbitrary editable display text.
- [ ] Dragging moves objects and the lower-right handle resizes them.
- [ ] Double-clicking a text object edits it inline.
- [ ] X/Y/W/H fields update the selected object.
- [ ] The contextual toolbar follows selection and reflects the object type.
- [ ] Guided headline, offer-title, and response-copy fields update their bound canvas elements immediately.
- [ ] Templates are starting points, not locked compositions.
- [ ] The palette includes exactly Text, Image, QR code, Rectangle, Square, Circle, and Squircle.
- [ ] Preview/proof, download, and print-service actions remain semantically distinct.
- [ ] Autosave cannot publish, export, place an order, or charge.
- [ ] The application is usable at 320, 360, 736/768, 1024, and 1440px without horizontal overflow.
- [ ] Dark mode changes the app shell but not the literal print design.
- [ ] qurl remains the QR renderer and Heard remains the destination owner.

## 22. Continuation rules for the next design iteration

This document records the current mockup, not a claim that every detail is final. Future design feedback must be applied by editing this document and the visual mockup together.

When the user requests another iteration:

1. read this file and `docs/slice7.md` before editing;
2. identify the exact state, control, or interaction being changed;
3. preserve all unrelated settings and direct-edit capabilities;
4. update the visual mockup;
5. update the corresponding sections, inventories, states, and acceptance checks here;
6. record any new unresolved product decision explicitly;
7. do not begin production implementation until D0 is approved.

This keeps the visual artifact and the durable reconstruction contract synchronized even if a different model or a new task continues the work.
