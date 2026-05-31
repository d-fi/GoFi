# Changelog

## 2.3.2 - 2026-05-31

This release continues the 2.3.x follow-up work around save layouts, release dates, cover artwork, and web UI polish.

### Changed

- Aligned metadata release-date tagging with save layout release-date selection.
- Metadata tags now prefer richer album dates when available, including original and physical release dates, before falling back to Deezer's public album release date.
- The web layout Fields modal now shows sample values for common fields when a preview is loaded.
- Shared struct-to-map conversion is now centralized in `utils`.

### Fixed

- Fixed separate cover-file placement for single-disc albums when the save layout contains `{DISK_FOLDER}`. Covers now stay inside the album folder instead of moving up to the artist folder.
- Fixed public Deezer API HTTP errors being cached as valid responses.
- Fixed failed Deezer cover CDN responses being cached or saved as artwork.
- Added `ORIGINAL_RELEASE_DATE` decoding from Deezer album data so release date fields can use it when Deezer provides it.

## 2.3.1 - 2026-05-30

This release is a focused follow-up to 2.3.0. It improves compatibility with current Deezer links and playlist responses, adds more flexible save layouts for multi-disc albums, and tightens cover artwork handling in the CLI and web UI.

### Added

- Added support for modern Deezer share links such as `https://link.deezer.com/s/...`.
- Added fallback save layout placeholders using `{FIRST|SECOND|THIRD}` syntax.
- Added opt-in multi-disc folder layouts with `{DISK_FOLDER}` and `{DISK_NUMBER}`.
- Added support for custom cover sizes between `50` and `1800`.
- Added web handling for custom saved cover sizes, shown as custom dropdown options.

### Changed

- Kept the existing default multi-disc album behavior unchanged unless `{DISK_FOLDER}` is used in the save layout.
- Changed separate cover-file saving for `{DISK_FOLDER}` layouts so one cover file is saved at the album root instead of once per disc folder.
- Improved release date placeholder fallback behavior for layouts.
- Improved web preview behavior by removing redundant success toasts when tracks are already shown in the preview table.
- Improved release package checks so package contents are verified after `make pkg`.

### Fixed

- Fixed Deezer playlist decoding when `ALB_ID`, `ART_ID`, `SNG_ID`, and related IDs are returned as numbers instead of strings.
- Fixed request cache collisions by including request params in cache keys.
- Fixed cover size validation so web-selected high-resolution values like `1200` and `1400` do not fail during download.
- Fixed invalid manual cover sizes so out-of-range values fall back to the existing/default value instead of being saved.
- Fixed layout placeholder parsing for fallback placeholders and disc-folder detection.

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
