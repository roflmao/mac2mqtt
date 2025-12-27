# Changelog for mac2mqtt

```
2.0.1   2025-12-27
        [Patch]
        * Update README with comprehensive sensor and command documentation
        * Fix display sleep to work when running as LaunchDaemon
        * [skip ci] Clean up CHANGELOG - remove auto-generated 1.0.12 entry
        * [skip ci] Prevent auto-increment when manual version tag exists
        * [skip ci] Update CHANGELOG for v1.0.12

2.0.0   2025-12-27
        [Major]
        * Add system uptime monitoring with relative time display in Home Assistant
        * Fix auto-update to check correct repository (roflmao/mac2mqtt instead of bessarabov/mac2mqtt)
        * Fix auto-update architecture detection to download native ARM64 binaries on Apple Silicon
        * Improve LaunchDaemon documentation with correct launchctl commands
        * Fix uptime sensor display formatting in Home Assistant

1.0.9   2025-12-27
        [Patch]
        * Clarify launchctl commands for LaunchDaemon management

1.0.8   2025-12-27
        [Patch]
        * Fix auto-update to detect native system architecture

1.0.7   2025-12-27
        [Patch]
        * Add system uptime monitoring

1.0.6   2025-12-27
        [Patch]
        * Improve installation and LaunchDaemon documentation

1.0.5   2025-12-27
        [Patch]
        * Fix auto-update to check correct repository
        * Update README URLs [skip ci]

1.0.4   2025-12-21
        [Patch]
        * Fix changelog workflow indentation
        * Automate changelog updates in release pipeline

1.4.0   2024-03-10
        [Minor]
        * Added go.mod, go.sum and new GitHub Action to create binaries for Apple Silicon Macs & Intel-based Macs

1.3.1   2021-04-24
        [Patch]
        * Sending battery state every 60 seconds instead of 1

1.3.0   2021-04-24
        [Minor]
        * Sending information about battery charge percent to MQTT topic PREFIX + /status/battery

1.2.0   2021-04-13
        [Minor]
        * New command /command/shutdown

1.1.0   2021-04-12
        [Minor]
        * New command /command/displaysleep to turn off mac display

1.0.0   2021-04-12
        [Major]
        * First public release
```
