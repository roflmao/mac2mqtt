# mac2mqtt

Control and monitor your macOS computer via MQTT. Integrates seamlessly with Home Assistant and other home automation systems.

Originally created by [Ivan Bessarabov](https://github.com/bessarabov/mac2mqtt).

## Features

### Status Monitoring

The following information is published to MQTT:

* Current volume level (0-100)
* Volume mute state (true/false)
* Battery charge percentage
* Connection status (alive/offline)
* Active application name
* Wi-Fi network name (SSID)
* Wi-Fi signal strength (RSSI)
* Wi-Fi IP address
* Network upload rate (KB/s)
* Network download rate (KB/s)
* System uptime

### Remote Control

Send commands via MQTT to control your Mac:

| Command | Topic | Accepted Values | Description |
|---------|-------|-----------------|-------------|
| **Set Volume** | `mac2mqtt/HOSTNAME/command/volume` | `0` - `100` | Set system volume level |
| **Mute/Unmute** | `mac2mqtt/HOSTNAME/command/mute` | `true` / `false` | Mute or unmute system audio |
| **Sleep** | `mac2mqtt/HOSTNAME/command/sleep` | `sleep` | Put computer to sleep |
| **Shutdown** | `mac2mqtt/HOSTNAME/command/shutdown` | `shutdown` | Shut down computer |
| **Display Sleep** | `mac2mqtt/HOSTNAME/command/displaysleep` | `displaysleep` | Turn off display only |

### Home Assistant Integration

* **MQTT Auto Discovery** - Automatically creates all entities in Home Assistant
* **Device grouping** - All controls grouped under one device
* **Availability tracking** - Shows online/offline status

## Available Metrics

mac2mqtt publishes the following metrics to MQTT:

| Metric | Topic | Values | Update Frequency | Description |
|--------|-------|--------|------------------|-------------|
| **Connection Status** | `mac2mqtt/HOSTNAME/status/alive` | `true` / `false` | On connect/disconnect | Indicates if mac2mqtt is connected to MQTT broker (uses Last Will and Testament) |
| **Volume Level** | `mac2mqtt/HOSTNAME/status/volume` | `0` - `100` | Every 2 seconds | Current system volume level as percentage |
| **Mute Status** | `mac2mqtt/HOSTNAME/status/mute` | `true` / `false` | Every 2 seconds | Whether system audio is muted |
| **Battery Charge** | `mac2mqtt/HOSTNAME/status/battery` | `0` - `100` | Every 60 seconds | Battery charge percentage (laptops only) |
| **Active Application** | `mac2mqtt/HOSTNAME/status/active_app` | String | Every 2 seconds | Name of the currently active (frontmost) application |
| **Wi-Fi SSID** | `mac2mqtt/HOSTNAME/status/wifi_ssid` | String | Every 60 seconds | Name of the currently connected Wi-Fi network |
| **Wi-Fi Signal Strength** | `mac2mqtt/HOSTNAME/status/wifi_signal_strength` | dBm value | Every 60 seconds | Wi-Fi signal strength (RSSI value, typically -30 to -90) |
| **Wi-Fi IP Address** | `mac2mqtt/HOSTNAME/status/wifi_ip` | IPv4 address | Every 60 seconds | Current IPv4 address of the primary Wi-Fi interface (en0) |
| **Last Boot Time** | `mac2mqtt/HOSTNAME/status/uptime` | ISO 8601 timestamp | Every 60 seconds | Timestamp of when the system last booted (displays as relative time in Home Assistant) |
| **Network Upload Rate** | `mac2mqtt/HOSTNAME/status/network_upload_rate` | KB/s (decimal) | Every 2 seconds | Current upload rate in kilobytes per second |
| **Network Download Rate** | `mac2mqtt/HOSTNAME/status/network_download_rate` | KB/s (decimal) | Every 2 seconds | Current download rate in kilobytes per second |

**Notes:**
- `HOSTNAME` is automatically derived from your macOS computer's hostname (e.g., `bessarabov-osx`)
- **Wi-Fi SSID is unavailable on macOS Ventura+** due to privacy restrictions requiring Apple Developer certificate and code signing. The sensor will show "Not Connected". Signal strength and IP address work without restrictions.

All metrics are published immediately upon connection and then updated according to their schedules.

## Requirements

* macOS (any version with `osascript` and `pmset` support)
* MQTT broker (e.g., Mosquitto)
* Go 1.22+ (only if building from source)

## Overview

To use mac2mqtt, you need:

* **mac2mqtt binary** - The compiled executable
* **Configuration file** - `mac2mqtt.yaml` with your MQTT broker settings
* **MQTT broker** - A running MQTT server (local or remote)
* **Optional: Automation system** - Home Assistant, Node-RED, or any MQTT-compatible platform

Originally created for [Home Assistant](https://www.home-assistant.io/) integration, mac2mqtt works with any MQTT-compatible automation system.

### Recommended Directory Structure

**Option 1: System-wide (recommended for LaunchDaemon):**

```
/usr/local/mac2mqtt/
├── mac2mqtt
└── mac2mqtt.yaml
```

**Option 2: User home directory:**

```
/Users/USERNAME/mac2mqtt/
├── mac2mqtt
└── mac2mqtt.yaml
```

For running as a system service (LaunchDaemon), Option 1 is recommended as it doesn't require username customization in the plist file.

## Installation

### Option 1: Download Pre-compiled Binary

1. Go to the [Releases](../../releases) page
2. Download the correct file for your Mac:
   * `mac2mqtt_VERSION_arm64` - Apple Silicon (M1, M2, M3, M4)
   * `mac2mqtt_VERSION_x86_64` - Intel-based Macs
3. Create the directory and prepare the binary:
   ```bash
   mkdir -p ~/mac2mqtt
   cd ~/Downloads
   chmod +x mac2mqtt_VERSION_ARCH
   xattr -d com.apple.quarantine mac2mqtt_VERSION_ARCH  # Remove macOS quarantine
   mv mac2mqtt_VERSION_ARCH ~/mac2mqtt/mac2mqtt
   ```

**Important:** The `xattr -d com.apple.quarantine` command is required to remove the macOS security quarantine flag from downloaded files. Without this step, macOS will prevent the binary from running.

### Option 2: Build from Source

Requirements: [Go 1.22+](https://go.dev/doc/install)

```bash
git clone https://github.com/roflmao/mac2mqtt.git
cd mac2mqtt
go build .
```

This creates the `mac2mqtt` executable in the current directory.

## Configuration

1. Create the configuration file `mac2mqtt.yaml`:

```yaml
# MQTT broker settings
mqtt_ip: 192.168.1.123
mqtt_port: 1883
mqtt_user: your_username
mqtt_password: your_password

# Debug mode - logs all MQTT publishes (optional, default: false)
debug: false

# Dry-run mode - simulates MQTT without actual connection (optional, default: false)
# Useful for testing without a real MQTT broker
dry_run: false
```

2. Edit the values to match your MQTT broker settings

### Configuration Options

* **mqtt_ip** (required*) - IP address of your MQTT broker (*not required in dry-run mode)
* **mqtt_port** (required*) - Port of your MQTT broker, usually 1883 (*not required in dry-run mode)
* **mqtt_user** (optional) - Username for MQTT authentication
* **mqtt_password** (optional) - Password for MQTT authentication
* **debug** (optional) - Enable debug logging to see all MQTT messages being published (default: false)
* **dry_run** (optional) - Test mode that simulates MQTT without connecting to a real broker (default: false)
* **auto_update** (optional) - Enable automatic updates from GitHub releases (default: true)

#### Debug Mode

When `debug: true`, you'll see detailed logs like:
```
[DEBUG] Publishing to topic 'mac2mqtt/your-mac/status/volume': 50 (QoS=0, Retained=false)
[DEBUG] Publishing to topic 'mac2mqtt/your-mac/status/mute': false (QoS=0, Retained=false)
```

#### Dry-Run Mode

When `dry_run: true`, mac2mqtt will simulate all operations without connecting to an actual MQTT broker. This is useful for:
- Testing configuration without an MQTT server
- Seeing what topics and payloads would be published
- Development and debugging

In dry-run mode, you'll see logs like:
```
DRY RUN MODE ENABLED - No actual MQTT connection will be made
Dry-run mode: Simulating MQTT connection
Connected to MQTT
[DRY-RUN] Publishing to topic 'mac2mqtt/your-mac/status/alive': true (QoS=0, Retained=true)
[DRY-RUN] Publishing to topic 'homeassistant/binary_sensor/mac2mqtt_your-mac/your-mac_alive/config': {...}
[DRY-RUN] Publishing to topic 'mac2mqtt/your-mac/status/volume': 50 (QoS=0, Retained=false)
```

## Auto-Update

mac2mqtt includes automatic update functionality that keeps your installation up-to-date with the latest releases from GitHub.

### How It Works

By default, mac2mqtt automatically:
1. Checks for new releases every hour
2. Downloads the appropriate binary for your platform (darwin/arm64, darwin/amd64, etc.)
3. Validates the download integrity
4. Atomically replaces the current binary with a backup
5. Exits cleanly to allow launchd to restart with the new version

### Configuration

Auto-update is **enabled by default**. To disable it, add to your `mac2mqtt.yaml`:

```yaml
# Auto-update settings
auto_update: false  # Set to false to disable automatic updates
```

### Logging

When auto-update runs, you'll see logs like:

**Startup:**
```
Started
Version: v1.0.2
Auto-update enabled (checks every hour)
```

**Update available:**
```
Checking for updates...
New version available: v1.0.2 -> v1.0.3
Downloading update from https://github.com/roflmao/mac2mqtt/releases/download/v1.0.3/mac2mqtt-darwin-arm64 (15234567 bytes)
Downloaded 15234567 bytes to /tmp/mac2mqtt-update-123456
Replacing binary at /Users/USERNAME/mac2mqtt/mac2mqtt
Binary replaced successfully, backup saved to /Users/USERNAME/mac2mqtt/mac2mqtt.old
Update to v1.0.3 completed successfully
Exiting to allow launchd restart with new binary
```

**No update needed:**
```
Checking for updates...
Already running latest version (current: v1.0.3, latest: v1.0.3)
No update required
```

### Requirements for Auto-Update

**For launchd service users** (recommended):
- Ensure `KeepAlive` is set to `true` in your plist file (see Running as a Background Service section)
- mac2mqtt needs write permissions to its binary location

**For manual execution:**
- Updates will download and install, but you'll need to manually restart the process

### Rollback

If an update causes issues, you can rollback to the previous version:

```bash
# Stop the service
sudo launchctl unload /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist

# Restore the backup
mv /Users/USERNAME/mac2mqtt/mac2mqtt.old /Users/USERNAME/mac2mqtt/mac2mqtt

# Restart the service
sudo launchctl load /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
```

Or disable auto-update by setting `auto_update: false` in your config file.

### Troubleshooting Auto-Update

**Permission errors:**
```
Update failed: insufficient permissions to update: cannot write to /Users/USERNAME/mac2mqtt
```

Solution: Ensure the user running mac2mqtt has write permissions to the binary location.

**Network errors:**
```
Update failed: failed to fetch release info: context deadline exceeded (will retry in 1 hour)
```

Solution: Check your internet connection and ensure GitHub is accessible. Updates will automatically retry every hour.

**Platform not found:**
```
Update failed: no binary found for platform darwin/arm64
```

Solution: This typically means a release hasn't been fully published yet. Updates will retry automatically.

### Security

- Downloads use HTTPS from GitHub releases
- Download size is validated against expected size
- Binary updates are atomic (all-or-nothing)
- A backup (.old) is created before replacement
- On failure, the system automatically rolls back to the previous version

## Running

### Manual Execution

Place both files in the same directory:

```
/Users/USERNAME/mac2mqtt/
├── mac2mqtt
└── mac2mqtt.yaml
```

Then run:

```bash
cd ~/mac2mqtt
./mac2mqtt
```

You should see output similar to:

```
$ ./mac2mqtt
2021/04/12 10:37:28 Started
2021/04/12 10:37:29 Connected to MQTT
2021/04/12 10:37:29 Sending 'true' to topic: mac2mqtt/bessarabov-osx/status/alive
```

### Running as a Background Service (LaunchDaemon)

To automatically start mac2mqtt on system boot, you'll install it as a system service using launchd.

#### Installation Options

**Option A: System-wide installation (Recommended)**

This installs mac2mqtt in `/usr/local/mac2mqtt/` - no username customization needed.

```bash
# Create the system directory and copy files
sudo mkdir -p /usr/local/mac2mqtt
sudo cp mac2mqtt /usr/local/mac2mqtt/
sudo cp mac2mqtt.yaml /usr/local/mac2mqtt/
sudo chmod +x /usr/local/mac2mqtt/mac2mqtt

# Copy the plist template (no customization needed!)
sudo cp com.bessarabov.mac2mqtt.plist.template /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
sudo chown root:wheel /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
sudo chmod 644 /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist

# Load the service
sudo launchctl load /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
```

**Option B: User home directory installation**

Keep files in your home directory (e.g., `~/mac2mqtt/`). You'll need to customize the plist:

```bash
# Edit the plist template
sudo cp com.bessarabov.mac2mqtt.plist.template /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
sudo nano /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
# Change /usr/local/mac2mqtt/ to /Users/YOUR_USERNAME/mac2mqtt/ in both Program and WorkingDirectory
```

Or create the file manually at `/Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>Label</key>
        <string>com.bessarabov.mac2mqtt</string>
        <key>Program</key>
        <string>/Users/USERNAME/mac2mqtt/mac2mqtt</string>
        <key>WorkingDirectory</key>
        <string>/Users/USERNAME/mac2mqtt/</string>
        <key>StandardOutPath</key>
        <string>/tmp/mac2mqtt.log</string>
        <key>StandardErrorPath</key>
        <string>/tmp/mac2mqtt.err</string>
        <key>RunAtLoad</key>
        <true/>
        <key>KeepAlive</key>
        <true/>
    </dict>
</plist>
```

**Important:** Replace `USERNAME` with your actual macOS username.

**Configuration Notes:**
- `RunAtLoad` - Service starts automatically at boot
- `KeepAlive` - Service automatically restarts if it crashes
- `StandardOutPath` and `StandardErrorPath` - Log files for debugging (optional but recommended)

3. Set correct permissions and load the service:

```bash
sudo chown root:wheel /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
sudo chmod 644 /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
sudo launchctl load /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
```

4. Managing the service:

**IMPORTANT:** Do NOT use `launchctl stop` or `launchctl start` - they don't work with LaunchDaemons that have `KeepAlive: true`. Use the commands below instead.

**Correct commands:**

```bash
# Restart the service (most common operation)
sudo launchctl kickstart -k system/com.bessarabov.mac2mqtt

# Stop the service completely (unload it)
sudo launchctl unload /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist

# Start the service (after unloading)
sudo launchctl load /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist

# Check if service is running
sudo launchctl list | grep mac2mqtt

# View detailed service status
sudo launchctl print system/com.bessarabov.mac2mqtt

# View logs (if logging is enabled in plist)
tail -f /tmp/mac2mqtt.err
tail -f /tmp/mac2mqtt.log
```

**Why `stop` and `start` don't work:**
- `launchctl stop` - The service immediately restarts because `KeepAlive: true`
- `launchctl start` - Only works for LaunchAgents, not LaunchDaemons

## Home Assistant Integration

### MQTT Auto Discovery

mac2mqtt supports Home Assistant's MQTT discovery feature. When mac2mqtt connects to your MQTT broker, it automatically publishes discovery messages that Home Assistant will detect.

**Automatically discovered entities:**

* Binary Sensor - Status (connection status)
* Sensor - Battery
* Sensor - Volume Level (read-only)
* Sensor - Active App
* Sensor - Wi-Fi SSID
* Sensor - Wi-Fi Signal Strength
* Sensor - Wi-Fi IP
* Sensor - Last Boot (timestamp)
* Sensor - Network Upload (KB/s)
* Sensor - Network Download (KB/s)
* Switch - Mute
* Number - Volume (0-100)
* Button - Sleep
* Button - Shutdown
* Button - Display Sleep

All entities are grouped under a single device in Home Assistant using your computer's hostname.

**No manual configuration required!** Simply ensure:
1. MQTT integration is enabled in Home Assistant
2. MQTT discovery is enabled (it's on by default)
3. mac2mqtt is running and connected to your MQTT broker

The entities will automatically appear in Home Assistant and can be added to your dashboard.

![Home Assistant Example](https://user-images.githubusercontent.com/47263/114361105-753c4200-9b7e-11eb-833c-c26a2b7d0e00.png)

### Manual Configuration (Optional)

If you prefer manual configuration or need custom scripts, you can still configure entities manually:

#### configuration.yaml

```yaml
script:
  air2_sleep:
    icon: mdi:laptop
    sequence:
      - service: mqtt.publish
        data:
          topic: "mac2mqtt/bessarabov-osx/command/sleep"
          payload: "sleep"

  air2_shutdown:
    icon: mdi:laptop
    sequence:
      - service: mqtt.publish
        data:
          topic: "mac2mqtt/bessarabov-osx/command/shutdown"
          payload: "shutdown"

  air2_displaysleep:
    icon: mdi:laptop
    sequence:
      - service: mqtt.publish
        data:
          topic: "mac2mqtt/bessarabov-osx/command/displaysleep"
          payload: "displaysleep"

sensor:
  - platform: mqtt
    name: air2_alive
    icon: mdi:laptop
    state_topic: "mac2mqtt/bessarabov-osx/status/alive"

  - platform: mqtt
    name: "air2_battery"
    icon: mdi:battery-high
    unit_of_measurement: "%"
    state_topic: "mac2mqtt/bessarabov-osx/status/battery"

switch:
  - platform: mqtt
    name: air2_mute
    icon: mdi:volume-mute
    state_topic: "mac2mqtt/bessarabov-osx/status/mute"
    command_topic: "mac2mqtt/bessarabov-osx/command/mute"
    payload_on: "true"
    payload_off: "false"

number:
  - platform: mqtt
    name: air2_volume
    icon: mdi:volume-medium
    state_topic: "mac2mqtt/bessarabov-osx/status/volume"
    command_topic: "mac2mqtt/bessarabov-osx/command/volume"
```

### ui-lovelace.yaml

```yaml
title: Home
views:
  - path: default_view
    title: Home
    cards:
      - type: entities
        entities:
          - sensor.air2_alive
          - sensor.air2_battery
          - type: 'custom:slider-entity-row'
            entity: number.air2_volume
            min: 0
            max: 100
          - switch.air2_mute
          - type: button
            name: air2
            entity: script.air2_sleep
            action_name: sleep
            tap_action:
              action: call-service
              service: script.air2_sleep
          - type: button
            name: air2
            entity: script.air2_shutdown
            action_name: shutdown
            tap_action:
              action: call-service
              service: script.air2_shutdown
          - type: button
            name: air2
            entity: script.air2_displaysleep
            action_name: displaysleep
            tap_action:
              action: call-service
              service: script.air2_displaysleep

      - type: history-graph
        hours_to_show: 48
        refresh_interval: 0
        entities:
          - sensor.air2_battery
```

**Note:** Replace `bessarabov-osx` with your computer's hostname in all topic paths.

## MQTT Topics Reference

All topics use the format: `mac2mqtt/COMPUTER_NAME/status/#` or `mac2mqtt/COMPUTER_NAME/command/#`

The `COMPUTER_NAME` is automatically derived from your macOS hostname.

### Status Topics

mac2mqtt publishes to these topics:

#### `mac2mqtt/COMPUTER_NAME/status/alive`

**Values:** `true` or `false`

Indicates connection status using MQTT Last Will and Testament:
* `true` - mac2mqtt is connected to the MQTT broker
* `false` - mac2mqtt is disconnected

#### `mac2mqtt/COMPUTER_NAME/status/volume`

**Values:** `0` to `100`

Current volume level of the computer.

**Update frequency:** Every 2 seconds

#### `mac2mqtt/COMPUTER_NAME/status/mute`

**Values:** `true` or `false`

Mute status:
* `true` - Computer is muted (no sound)
* `false` - Computer is not muted

**Update frequency:** Every 2 seconds

#### `mac2mqtt/COMPUTER_NAME/status/battery`

**Values:** `0` to `100`

Battery charge percentage.

**Update frequency:** Every 60 seconds

#### `mac2mqtt/COMPUTER_NAME/status/active_app`

**Values:** String (application name)

Name of the currently active (frontmost) application.

**Update frequency:** Every 2 seconds

**Example values:** `Safari`, `Terminal`, `Visual Studio Code`

#### `mac2mqtt/COMPUTER_NAME/status/wifi_ssid`

**Values:** String (network name) or `Not Connected`

Name of the currently connected Wi-Fi network.

**Update frequency:** Every 60 seconds

**Important:** On macOS Ventura and later, accessing Wi-Fi SSID requires a properly signed application with Apple Developer certificate and entitlements. The current command-line version will show "Not Connected" regardless of actual connection status. Signal strength and IP address are not affected by this limitation.

#### `mac2mqtt/COMPUTER_NAME/status/wifi_signal_strength`

**Values:** Integer (dBm)

Wi-Fi signal strength in dBm (typically -30 to -90).

**Update frequency:** Every 60 seconds

**Range:** Higher (less negative) values = stronger signal. Example: -40 dBm is stronger than -70 dBm.

#### `mac2mqtt/COMPUTER_NAME/status/wifi_ip`

**Values:** IPv4 address or `Not Connected`

Current IPv4 address of the primary Wi-Fi interface.

**Update frequency:** Every 60 seconds

**Example values:** `192.168.1.100`, `10.0.0.50`

#### `mac2mqtt/COMPUTER_NAME/status/uptime`

**Values:** ISO 8601 timestamp

Timestamp of when the system last booted.

**Update frequency:** Every 60 seconds

**Example values:** `2025-12-27T19:10:18Z`

**Note:** Home Assistant displays this as relative time (e.g., "2 hours ago").

#### `mac2mqtt/COMPUTER_NAME/status/network_upload_rate`

**Values:** Decimal number (KB/s)

Current network upload rate in kilobytes per second.

**Update frequency:** Every 2 seconds

**Example values:** `0.00`, `12.45`, `156.78`

**Note:** Rate is calculated by comparing network interface statistics over time. First measurement will always be `0.00`.

#### `mac2mqtt/COMPUTER_NAME/status/network_download_rate`

**Values:** Decimal number (KB/s)

Current network download rate in kilobytes per second.

**Update frequency:** Every 2 seconds

**Example values:** `0.00`, `234.56`, `1024.32`

**Note:** Rate is calculated by comparing network interface statistics over time. First measurement will always be `0.00`.

### Command Topics

Send messages to these topics to control your Mac:

#### `mac2mqtt/COMPUTER_NAME/command/volume`

**Values:** `0` to `100`

Set the computer volume level.

**Example:**
```bash
mosquitto_pub -t "mac2mqtt/your-mac/command/volume" -m "50"
```

#### `mac2mqtt/COMPUTER_NAME/command/mute`

**Values:** `true` or `false`

Control mute state:
* `true` - Mute the computer
* `false` - Unmute the computer

**Example:**
```bash
mosquitto_pub -t "mac2mqtt/your-mac/command/mute" -m "true"
```

#### `mac2mqtt/COMPUTER_NAME/command/sleep`

**Value:** `sleep`

Put the computer into sleep mode.

**Example:**
```bash
mosquitto_pub -t "mac2mqtt/your-mac/command/sleep" -m "sleep"
```

#### `mac2mqtt/COMPUTER_NAME/command/shutdown`

**Value:** `shutdown`

Shut down the computer.

**Behavior:**
* If run as `root` - Always shuts down
* If run as regular user - May fail if other users are logged in

**Example:**
```bash
mosquitto_pub -t "mac2mqtt/your-mac/command/shutdown" -m "shutdown"
```

#### `mac2mqtt/COMPUTER_NAME/command/displaysleep`

**Value:** `displaysleep`

Turn off the display.

**Example:**
```bash
mosquitto_pub -t "mac2mqtt/your-mac/command/displaysleep" -m "displaysleep"
```

## Troubleshooting

### LaunchDaemon Service Won't Start

If the LaunchDaemon fails to start (exit code 78 or similar):

1. **Check execute permissions** on the binary:
   ```bash
   # For system-wide installation
   ls -la /usr/local/mac2mqtt/mac2mqtt
   # OR for user directory installation
   ls -la ~/mac2mqtt/mac2mqtt
   ```
   Should show `-rwxr-xr-x`. If not, fix with:
   ```bash
   # For system-wide installation
   sudo chmod +x /usr/local/mac2mqtt/mac2mqtt
   # OR for user directory installation
   chmod +x ~/mac2mqtt/mac2mqtt
   ```

2. **Remove macOS quarantine flag** from downloaded binaries:
   ```bash
   # For system-wide installation
   xattr -l /usr/local/mac2mqtt/mac2mqtt
   # OR for user directory installation
   xattr -l ~/mac2mqtt/mac2mqtt
   ```
   If you see `com.apple.quarantine`, remove it:
   ```bash
   # For system-wide installation
   sudo xattr -d com.apple.quarantine /usr/local/mac2mqtt/mac2mqtt
   # OR for user directory installation
   xattr -d com.apple.quarantine ~/mac2mqtt/mac2mqtt
   ```

3. **Check service status**:
   ```bash
   sudo launchctl print system/com.bessarabov.mac2mqtt
   ```
   Look for `last exit code` and `state` fields.

4. **View logs** (if StandardOutPath/StandardErrorPath are configured):
   ```bash
   cat /tmp/mac2mqtt.err
   cat /tmp/mac2mqtt.log
   ```

5. **Restart the service** after fixing issues:
   ```bash
   sudo launchctl kickstart -k system/com.bessarabov.mac2mqtt
   ```

### Connection Issues

If mac2mqtt cannot connect to MQTT:

1. Verify your MQTT broker is running
2. Check `mac2mqtt.yaml` settings (IP, port, credentials)
3. Test connectivity: `ping YOUR_MQTT_IP`
4. Verify firewall settings allow MQTT traffic (default port 1883)
5. Check logs for connection errors

### Permission Issues

If commands don't work:

1. Ensure Terminal/mac2mqtt has accessibility permissions
2. Go to System Settings > Privacy & Security > Accessibility
3. Add Terminal or the mac2mqtt binary to the allowed list

### Binary Won't Execute (Killed: 9)

If you see `Killed: 9` when trying to run the binary:

1. This is macOS Gatekeeper blocking the downloaded file
2. Remove the quarantine attribute:
   ```bash
   xattr -d com.apple.quarantine ~/mac2mqtt/mac2mqtt
   ```
3. Alternatively, you can right-click the binary in Finder and select "Open" to approve it

### Finding Your Computer Name

To see the exact topic prefix mac2mqtt will use:

```bash
hostname | cut -d. -f1 | tr -cd 'a-zA-Z0-9_-'
```

## License

MIT

## Author

Created by [Ivan Bessarabov](https://github.com/bessarabov)
