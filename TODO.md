# TODO - Feature Roadmap

This document tracks potential features and enhancements for mac2mqtt.

## Current Features

### ✅ Implemented
- [x] Battery status monitoring
- [x] Volume level monitoring and control
- [x] Mute/unmute control
- [x] Online/connectivity status (alive)
- [x] Sleep mode control
- [x] Shutdown command
- [x] Display sleep control
- [x] Home Assistant MQTT auto-discovery
- [x] Active application tracking
- [x] Debug mode
- [x] Dry-run mode for testing
- [x] Exponential backoff retry on connection failure
- [x] Network activity monitoring (upload/download rates)
- [x] System uptime monitoring
- [x] Wi-Fi SSID monitoring
- [x] Wi-Fi signal strength monitoring
- [x] Wi-Fi IP address monitoring

## Planned Features

### Security & Connectivity

- [ ] **TLS/SSL Support**
  - Add TLS configuration options to `mac2mqtt.yaml`
  - Support for client certificates
  - Support for CA certificates
  - Enable secure connections to MQTT brokers

### System Monitoring

- [ ] **Disk Usage**
  - Total disk space
  - Used disk space
  - Free disk space
  - Per-partition monitoring

- [ ] **CPU Usage**
  - Current CPU usage percentage
  - Per-core usage (optional)

- [ ] **Memory Usage**
  - Total RAM
  - Used RAM
  - Free RAM
  - Memory pressure

- [ ] **Temperature Sensors**
  - CPU temperature
  - GPU temperature
  - Battery temperature
  - Fan speeds
  - Requires powermetrics or third-party tools

- [x] **Network Activity**
  - Upload rate
  - Download rate
  - Network interface statistics

- [x] **System Uptime**
  - Time since last boot

### Display

- [ ] **Brightness Control**
  - Get current brightness level
  - Set brightness level
  - Auto-brightness status

- [ ] **Display Resolution**
  - Current resolution
  - Available resolutions
  - External displays

### Audio Enhancements

- [ ] **Audio Input Device**
  - Current microphone/input device name
  - Available input devices

- [ ] **Audio Output Device**
  - Current speaker/output device name
  - Available output devices

- [ ] **Audio Device Switching**
  - Switch between audio input devices
  - Switch between audio output devices

### Power Management

- [ ] **Battery Enhancements**
  - Battery cycle count
  - Battery health percentage
  - Battery time remaining
  - Battery temperature
  - Is charging status

- [ ] **Power Source**
  - AC power vs battery
  - Power adapter wattage

- [ ] **Prevent Sleep (Caffeinate)**
  - Prevent system sleep
  - Prevent display sleep
  - Duration-based caffeinate

### Wi-Fi

- [ ] **Wi-Fi SSID**
  - Current network name

- [ ] **Wi-Fi Signal Strength**
  - RSSI value
  - Signal quality percentage

- [ ] **Wi-Fi IP Address**
  - Current IPv4 address
  - Current IPv6 address
  - Gateway address

- [ ] **Wi-Fi Control**
  - Turn Wi-Fi on/off
  - Connect to specific network

### Bluetooth

- [ ] **Bluetooth Status**
  - On/off state

- [ ] **Connected Devices**
  - List of connected Bluetooth devices
  - Device types (keyboard, mouse, headphones, etc.)

- [ ] **Bluetooth Control**
  - Turn Bluetooth on/off

### Privacy & Security

- [ ] **Camera In Use**
  - Detect when camera is active

- [ ] **Microphone In Use**
  - Detect when microphone is active

- [ ] **Screen Lock Status**
  - Locked/unlocked state

- [ ] **Lock Screen Command**
  - Trigger screen lock

### Focus & Notifications

- [ ] **Do Not Disturb**
  - Get DND status
  - Enable/disable DND

- [ ] **Focus Mode**
  - Current focus mode (Work, Personal, Sleep, etc.)
  - Set focus mode

- [ ] **Notifications**
  - Recent notifications
  - Notification count

### Media

- [ ] **Currently Playing Media**
  - Track name
  - Artist name
  - Album name
  - Source app (Music, Spotify, etc.)

- [ ] **Media Controls**
  - Play/pause
  - Next track
  - Previous track

### Process & Application

- [ ] **Running Applications**
  - List of all running applications
  - Application resource usage

- [ ] **Launch Application**
  - Open specific application by name

- [ ] **Quit Application**
  - Close specific application

### System Control

- [x] **Restart Command**
  - Restart the computer

- [ ] **Log Out Command**
  - Log out current user

### Configuration

- [ ] **Configurable Update Intervals**
  - Allow customization of update frequencies per metric
  - Different intervals for different sensors

- [ ] **Topic Prefix Customization**
  - Allow custom MQTT topic prefix instead of hardcoded "mac2mqtt"

- [ ] **Selective Metric Publishing**
  - Enable/disable specific metrics
  - Reduce unnecessary updates

## Implementation Priority

### Phase 1 - Core Improvements (High Priority)
1. TLS/SSL support
2. Disk usage monitoring
3. CPU usage
4. Memory usage
5. Configurable update intervals

### Phase 2 - System Monitoring (Medium Priority)
1. Temperature sensors
2. Network activity
3. Battery enhancements
4. System uptime
5. Power source detection

### Phase 3 - Network & Connectivity
1. Wi-Fi SSID and signal strength
2. Wi-Fi IP address
3. Bluetooth status
4. Connected Bluetooth devices

### Phase 4 - Advanced Controls
1. Brightness control
2. Wi-Fi management
3. Bluetooth management
4. DND/Focus control
5. Screen lock

### Phase 5 - Media & Applications
1. Currently playing media
2. Media controls
3. Running applications
4. Launch/quit application commands

## Technical Considerations

### macOS APIs & Tools
- `pmset` - Power management
- `osascript` - AppleScript for various controls
- `networksetup` - Network configuration
- `system_profiler` - System information
- `powermetrics` - Advanced power and temperature metrics (requires sudo)
- `sysctl` - System control variables
- `ioreg` - I/O Registry for hardware info

### Permissions Required
- Accessibility permissions for some AppleScript commands
- Full Disk Access for certain system information
- Screen Recording permission for camera detection
- Microphone permission for microphone detection

### Performance
- Avoid polling expensive operations too frequently
- Cache values where appropriate
- Consider implementing on-demand queries for expensive metrics

## Notes

- Maintain backward compatibility with existing topics
- Follow Home Assistant naming conventions for auto-discovery
- Prioritize features that work without requiring additional permissions
- Document all required permissions clearly
- Keep the application lightweight and efficient
