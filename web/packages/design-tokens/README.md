# Heard design-token snapshot

This release snapshot is copied verbatim from Haiven's central `@haiven/design-tokens` package while that package is not published to a registry.

Heard imports `src/tokens.css` directly so isolated application and Docker environments do not require a sibling repository or a package-manager link. When the central token release changes, update both files in `src/`. The brand-theme check compares them with the central source whenever it is available.
