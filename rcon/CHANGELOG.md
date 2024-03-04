# Changelog
All notable changes to this project will be documented in this file.

**ATTN**: This project uses [semantic versioning](http://semver.org/).

## [Unreleased]

## [v1.3.4] - 2022-11-12
### Fixed
- Minor fixes in packet package.

### Updated
- Updated Golang version to 1.19.
- Updated golangci linter to 1.50.1 version.

## [v1.3.3] - 2022-05-16
### Fixed
- Added "response from not rcon server" error on auth request (Re fixed panic: runtime error: makeslice: len out of range in rcon.Dial #5).

## [v1.3.2] - 2022-05-16
### Fixed
- Fixed panic: runtime error: makeslice: len out of range in rcon.Dial

### Updated
- Updated golangci linter to 1.42.1 version

## [v1.3.1] - 2021-01-06
### Updated
- Updated golangci linter to 1.33 version

### Changed
- Changed errors handling - added wrapping.

## [v1.3.0] - 2020-12-02
### Fixed
- Fixed wrong number of bytes written in Packet WriteTo function.

### Added
- Added rcontest Server for mocking RCON connections.

## [v1.2.4] - 2020-11-14
### Added
- Added the ability to run tests on a real Project Zomboid server. To do this, set environment variables 
`TEST_PZ_SERVER=true`, `TEST_PZ_SERVER_ADDR` and `TEST_PZ_SERVER_PASSWORD` with address and password from Project Zomboid
remote console.  
- Added the ability to run tests on a real Rust server. To do this, set environment variables `TEST_RUST_SERVER=true`, 
`TEST_RUST_SERVER_ADDR` and `TEST_RUST_SERVER_PASSWORD` with address and password from Rust remote console.  
- Added invalid padding test.

### Changed
- Changed CI workflows and related badges. Integration with Travis-CI was changed to GitHub actions workflow. Golangci-lint 
job was joined with tests workflow.  

## [v1.2.3] - 2020-10-20
### Fixed
- Fixed read/write deadline. The deadline was started from the moment the connection was established and was not updated 
after the command was sent.

## [v1.2.2] - 2020-10-18
### Added
- Added one more workaround for Rust server. When sent command "Say" there is no response data from server 
with packet.ID = SERVERDATA_EXECCOMMAND_ID, only previous console message that command was received with 
packet.ID = -1, therefore, forcibly set packet.ID to SERVERDATA_EXECCOMMAND_ID.

## [v1.2.1] - 2020-10-06
### Added
- Added authentication failed test.

### Changed
- Updated Golang version to 1.15.

## [v1.2.0] - 2020-07-10
### Added
- Added options to Dial. It is possible to set timeout and deadline settings.

### Fixed
- Change `SERVERDATA_AUTH_ID` and `SERVERDATA_EXECCOMMAND_ID` from 42 to 0. Conan Exiles has a bug because of which it 
always responds 42 regardless of the value of the request ID. This is no longer relevant, so the values have been 
changed.

### Changed
- Renamed `DefaultTimeout` const to `DefaultDeadline`
- Changed default timeouts from 10 seconds to 5 seconds

## [v1.1.2] - 2020-05-13
### Added
- Added go modules (go 1.13).
- Added golangci.yml linter config. To run linter use `golangci-lint run` command.
- Added CHANGELOG.md.
- Added more tests.

## v1.0.0 - 2019-07-27
### Added
- Initial implementation.

[Unreleased]: https://github.com/gorcon/rcon/compare/v1.3.4...HEAD
[v1.3.4]: https://github.com/gorcon/rcon/compare/v1.3.3...v1.3.4
[v1.3.3]: https://github.com/gorcon/rcon/compare/v1.3.2...v1.3.3
[v1.3.2]: https://github.com/gorcon/rcon/compare/v1.3.1...v1.3.2
[v1.3.1]: https://github.com/gorcon/rcon/compare/v1.3.0...v1.3.1
[v1.3.0]: https://github.com/gorcon/rcon/compare/v1.2.4...v1.3.0
[v1.2.4]: https://github.com/gorcon/rcon/compare/v1.2.3...v1.2.4
[v1.2.3]: https://github.com/gorcon/rcon/compare/v1.2.2...v1.2.3
[v1.2.2]: https://github.com/gorcon/rcon/compare/v1.2.1...v1.2.2
[v1.2.1]: https://github.com/gorcon/rcon/compare/v1.2.0...v1.2.1
[v1.2.0]: https://github.com/gorcon/rcon/compare/v1.1.2...v1.2.0
[v1.1.2]: https://github.com/gorcon/rcon/compare/v1.0.0...v1.1.2
