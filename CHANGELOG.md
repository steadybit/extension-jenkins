# Changelog

## v1.0.13

- Maintenance release

## v1.0.12

- chore: update to go 1.26.4
- feat: add weekly auto patch-release workflow

## v1.0.11

- Support discovery group attribute via `STEADYBIT_EXTENSION_DISCOVERY_GROUP` env var (or `discovery.group` Helm value) — when set, the extension adds `steadybit.group=<value>` to every discovered target
- Update dependencies

## v1.0.10

- Bump Go to 1.26.3
- Update dependencies

## v1.0.9

- Bump Go to 1.25.9
- Support if-none-match for the extension list endpoint
- Update dependencies

## v1.0.8

- feat(chart): split image.name into image.registry + image.name
- Support global.priorityClassName
- Update alpine packages in Docker image to address CVEs
- Update dependencies

## v1.0.7

- Update dependencies

## v1.0.6

- Update dependencies

## v1.0.5

- Update dependencies

## v1.0.4

- Updated dependencies

## v1.0.3

- Fix Job Start timeout

## v1.0.2

 - Added a Job Start timeout

## v1.0.1

 - Add support for self-signed certificates
 - Update dependencies

## v1.0.0

 - Initial release
