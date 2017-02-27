# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/) 
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [0.5.1] - 2017-02-27
### Changed
- SetGitPath(): better path handling

## [0.5.0] - 2017-02-27
### Added
- added `gitFetcher` type to ease mocking Git output while testing

## [0.4.0] - 2017-02-23
### Added
- `ReplaceHostWithIP` utility to perform DNS lookup and translate hosts in IPs
- now setting `gitPath` value is possible to use a different path for performing git commands (for example in a container where git is not available in the `PATH`)
- test coverage
### Changed
- `remoteTags` now returns error too

## [0.3.1] - 2017-02-19
### Added
- Apache v2.0 license

## [0.3.0] - 2017-02-19
### Added
- New `Version` API + tests
### Changed
- Reimplemented the `Version` semantic: using `github.com/Masterminds/semver` seemed overkill and was supporting only semver anyway, we want to be more generic and support versions like _1.0_ or _2.3.4.5_ too
- Version regexp (`versionRegexp`) is now more flexible and matches an arbitrary long digit dotted string (like _0.1.2.3.4.5.6..._ etc)
### Removed
- `github.com/Masterminds/semver` dependency

## [0.2.0] - 2017-02-11
### Added
- Added dependency to `github.com/Masterminds/semver` because I'm too lazy to implement a proper version ordering algorhitm. Unfortunately this lib is padding missing version numbers (i.e. minor and/or patch) with zeros, and I'd like to preseve the original ones from Git remotes, so maybe it will be replaced at some point.
- Added test `TestLatestVersion`
### Changed
- Bugfix: now thanks to `Masterminds/semver` we properly parse versions with double digits
- Small bugfix for `semverRegexp`

## [0.1.0] - 2017-02-10
### Added
- `repo.go` with first draft of utilities for parsing Git remote tags
- this changelog file to keep track of changes
