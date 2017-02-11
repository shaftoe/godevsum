# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/) 
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [0.2.0] - 2017-02-11
### Added
- Added dependency to `github.com/Masterminds/semver` because I'm too lazy to implement a proper version ordering algorhitm. Unfortunately this lib is padding missing version numbers (i.e. minor and/or patch) with zeros, and I'd like to preseve the original ones from Git remotes, so maybe it will be replaced at some point.
- Added test `TestLatestVersion`
### Modified
- Bugfix: now thanks to `Masterminds/semver` we properly parse versions with double digits
- Small bugfix for `semverRegexp`

## [0.1.0] - 2017-02-10
### Added
- `repo.go` with first draft of utilities for parsing Git remote tags
- this changelog file to keep track of changes
