# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.17] - 2023-06-29
- Fix incoming webhooks - use string for policy id (#47)

## [0.3.16] - 2023-05-01
- Fix docs for escalation policies

## [0.3.15] - 2023-01-25
- Add escalation policies (#41)
- Add email integrations (#42)
- Add incoming webhooks (#43)

## [0.3.14] - 2022-12-06
- Add automatic_reports, announcement_embed_visible, and read-only status_history to status page resources

## [0.3.13] - 2022-11-16
- Add more missing attributes
- Add custom_javascript attribute for status page resources

## [0.3.12] - 2022-05-18
- Add widget_type attribute for status page resources

## [0.3.11] - 2022-03-24
- Add policy_id attribute to the Heartbeat resource

## [0.3.10] - 2022-03-09
- Fix marshalling of empty array attributes in JSON (#24)

## [0.3.9] - 2022-03-04
- Add support for the `maintenance_timezone` attribute to the Monitor resource (#23)

## [0.3.8] - 2022-02-28
- Add support for the `history` attribute to the Status page resource (#22)   

## [0.3.7] - 2021-12-15
- Fix password behaviour on the Status page resource (#14)
- Add a computed url attributed for the Heartbeat resource (#17)
- Added follow_redirects and domain_expiration attributes to the Monitor resource (#18)

## [0.3.6] - 2021-12-08
- Added request_headers attribute to the Monitor resource (#13)
- Removed default subscribable=true value on the Status Page resource (#12)

## [0.3.2] - 2021-10-15
- Added Status Page Section resource

## [0.2.9] - 2021-08-30
- Added support for the `expected_status_code` monitor type

## [0.2.6] - 2021-08-13
- Updated documentation

## [0.2.5] - 2021-08-13
- Added read-only policy lookup

## [0.2.4] - 2021-08-12
- Initial release (migrated from https://github.com/BetterStackHQ/deprecated-terraform-provider-betteruptime)

[Unreleased]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.17...HEAD
[0.3.17]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.16...v0.3.17
[0.3.16]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.15...v0.3.16
[0.3.15]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.14...v0.3.15
[0.3.15]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.14...v0.3.15
[0.3.14]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.13...v0.3.14
[0.3.13]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.12...v0.3.13
[0.3.12]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.11...v0.3.12
[0.3.11]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.10...v0.3.11
[0.3.10]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.9...v0.3.10
[0.3.9]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.8...v0.3.9
[0.3.8]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.7...v0.3.8
[0.3.7]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.6...v0.3.7
[0.3.6]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.3.2...v0.3.6
[0.3.2]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.2.9...v0.3.2
[0.2.9]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.2.8...v0.2.9
[0.2.8]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.2.7...v0.2.8
[0.2.7]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.2.6...v0.2.7
[0.2.6]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/BetterStackHQ/terraform-provider-better-uptime/releases/tag/v0.2.4
