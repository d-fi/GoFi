# Changelog

## 2.3.0 - 2026-05-29

This release focuses on the web UI, Spotify conversion reliability, cover artwork handling, and release packaging.

### Added

- Added Linux ARM64, macOS ARM64, and Windows ARM64 release packages.
- Added a dark/light theme switch to the web UI, saved locally in the browser.
- Added explicit Save buttons for each web settings section.
- Added a collapsible Deezer ARL section in the web UI.
- Added a CLI-style track range selector to the web preview table.
- Added a web layout fields reference modal for save layout placeholders.
- Added configurable cover artwork behavior:
  - embed artwork in tracks
  - save artwork as a separate file
  - embed and save artwork
  - disable artwork
- Added configurable cover file names.
- Added release date layout placeholders including release date and release year.
- Added Spotify partner playlist fallback for playlist conversion.
- Added Spotify metadata matching fallback when ISRC matching is unavailable.

### Changed

- Improved Spotify playlist conversion and matching against Deezer.
- Switched Spotify matching to authenticated Deezer search.
- Tuned request retry behavior and converter concurrency.
- Improved Spotify match safety by rejecting mismatched featured artists and requiring the primary artist to match.
- Improved web preview flow so the Preview button sits beside the query input and Enter starts preview.
- Changed cover size inputs in the web UI to selectors with known working sizes.
- Disabled cover filename input when the selected cover mode does not write a separate file.
- Split the web UI into dedicated HTML, CSS, and JavaScript assets.
- Cleaned up web internals and removed stale request fields and handlers.
- Simplified cover metadata APIs and cover config normalization.
- Improved disabled field styling in the web UI.

### Fixed

- Fixed search result decoding when Deezer returns string booleans.
- Fixed Spotify matching edge cases around featured artists and version conflicts.
- Fixed release year parsing to use safer string splitting.
- Fixed stale/unnecessary web code after UI changes.

### Notes

- Existing package names remain for amd64 users:
  - `d-fi-linux.zip`
  - `d-fi-macos.zip`
  - `d-fi-win.zip`
- New ARM64 packages are:
  - `d-fi-linux-arm64.zip`
  - `d-fi-macos-arm64.zip`
  - `d-fi-win-arm64.zip`
