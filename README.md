# mac2mqtt

Control and monitor your macOS computer via MQTT. Integrates seamlessly with Home Assistant and other home automation systems.

## Features

### Status Monitoring

The following information is published to MQTT:

* Current volume level (0-100)
* Volume mute state (true/false)
* Battery charge percentage
* Connection status (alive/offline)

### Remote Control

Send commands via MQTT to:

* Set volume level
* Mute/unmute audio
* Put computer to sleep
* Shut down computer
* Turn off display

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

Place the executable and configuration file in your home directory:

```
/Users/USERNAME/mac2mqtt/
├── mac2mqtt
└── mac2mqtt.yaml
```

## Installation

### Option 1: Download Pre-compiled Binary

1. Go to the [Releases](../../releases) page
2. Download the correct file for your Mac:
   * `mac2mqtt_VERSION_arm64` - Apple Silicon (M1, M2, M3, M4)
   * `mac2mqtt_VERSION_x86_64` - Intel-based Macs
3. Make the file executable:
   ```bash
   chmod +x mac2mqtt_VERSION_ARCH
   mv mac2mqtt_VERSION_ARCH ~/mac2mqtt/mac2mqtt
   ```

### Option 2: Build from Source

Requirements: [Go 1.22+](https://go.dev/doc/install)

```bash
git clone https://github.com/bessarabov/mac2mqtt.git
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
```

2. Edit the values to match your MQTT broker settings

### Configuration Options

* **mqtt_ip** (required) - IP address of your MQTT broker
* **mqtt_port** (required) - Port of your MQTT broker (usually 1883)
* **mqtt_user** (optional) - Username for MQTT authentication
* **mqtt_password** (optional) - Password for MQTT authentication
* **debug** (optional) - Enable debug logging to see all MQTT messages being published (default: false)

When `debug: true`, you'll see detailed logs like:
```
[DEBUG] Publishing to topic 'mac2mqtt/your-mac/status/volume': 50 (QoS=0, Retained=false)
[DEBUG] Publishing to topic 'mac2mqtt/your-mac/status/mute': false (QoS=0, Retained=false)
```

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

### Running as a Background Service

To automatically start mac2mqtt on system boot:

1. Ensure `mac2mqtt.yaml` and `mac2mqtt` are in `/Users/USERNAME/mac2mqtt/`

2. Create the LaunchDaemon plist file at `/Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist`:

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
        <key>RunAtLoad</key>
        <true/>
        <key>KeepAlive</key>
        <true/>
    </dict>
</plist>
```

**Important:** Replace `USERNAME` with your actual macOS username.

3. Load the service:

```bash
sudo launchctl load /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
```

4. To stop the service:

```bash
sudo launchctl unload /Library/LaunchDaemons/com.bessarabov.mac2mqtt.plist
```

## Home Assistant Integration

![Home Assistant Example](https://user-images.githubusercontent.com/47263/114361105-753c4200-9b7e-11eb-833c-c26a2b7d0e00.png)

### configuration.yaml

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

### Connection Issues

If mac2mqtt cannot connect to MQTT:

1. Verify your MQTT broker is running
2. Check `mac2mqtt.yaml` settings (IP, port, credentials)
3. Test connectivity: `ping YOUR_MQTT_IP`
4. Verify firewall settings allow MQTT traffic (default port 1883)

### Permission Issues

If commands don't work:

1. Ensure Terminal/mac2mqtt has accessibility permissions
2. Go to System Settings > Privacy & Security > Accessibility
3. Add Terminal or the mac2mqtt binary to the allowed list

### Finding Your Computer Name

To see the exact topic prefix mac2mqtt will use:

```bash
hostname | cut -d. -f1 | tr -cd 'a-zA-Z0-9_-'
```

## License

MIT

## Author

Created by [Ivan Bessarabov](https://github.com/bessarabov)
