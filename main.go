package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var hostname string
var debugMode bool
var dryRunMode bool

type config struct {
	Ip       string `yaml:"mqtt_ip"`
	Port     string `yaml:"mqtt_port"`
	User     string `yaml:"mqtt_user"`
	Password string `yaml:"mqtt_password"`
	Debug    bool   `yaml:"debug"`
	DryRun   bool   `yaml:"dry_run"`
}

func (c *config) getConfig() *config {

	configContent, err := ioutil.ReadFile("mac2mqtt.yaml")
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(configContent, c)
	if err != nil {
		log.Fatal(err)
	}

	// Set global modes
	debugMode = c.Debug
	dryRunMode = c.DryRun

	if dryRunMode {
		log.Println("DRY RUN MODE ENABLED - No actual MQTT connection will be made")
	}

	// Only validate MQTT settings if not in dry run mode
	if !dryRunMode {
		if c.Ip == "" {
			log.Fatal("Must specify mqtt_ip in mac2mqtt.yaml")
		}

		if c.Port == "" {
			log.Fatal("Must specify mqtt_port in mac2mqtt.yaml")
		}
	}

	return c
}

func getHostname() string {

	hostname, err := os.Hostname()

	if err != nil {
		log.Fatal(err)
	}

	// "name.local" => "name"
	firstPart := strings.Split(hostname, ".")[0]

	// remove all symbols, but [a-zA-Z0-9_-]
	reg, err := regexp.Compile("[^a-zA-Z0-9_-]+")
	if err != nil {
		log.Fatal(err)
	}
	firstPart = reg.ReplaceAllString(firstPart, "")

	return firstPart
}

func getCommandOutput(name string, arg ...string) string {
	cmd := exec.Command(name, arg...)

	stdout, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	stdoutStr := string(stdout)
	stdoutStr = strings.TrimSuffix(stdoutStr, "\n")

	return stdoutStr
}

func getMuteStatus() bool {
	output := getCommandOutput("/usr/bin/osascript", "-e", "output muted of (get volume settings)")

	b, err := strconv.ParseBool(output)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func getCurrentVolume() int {
	output := getCommandOutput("/usr/bin/osascript", "-e", "output volume of (get volume settings)")

	i, err := strconv.Atoi(output)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func runCommand(name string, arg ...string) {
	cmd := exec.Command(name, arg...)

	_, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
}

// from 0 to 100
func setVolume(i int) {
	runCommand("/usr/bin/osascript", "-e", "set volume output volume "+strconv.Itoa(i))
}

// true - turn mute on
// false - turn mute off
func setMute(b bool) {
	runCommand("/usr/bin/osascript", "-e", "set volume output muted "+strconv.FormatBool(b))
}

func commandSleep() {
	runCommand("pmset", "sleepnow")
}

func commandDisplaySleep() {
	runCommand("pmset", "displaysleepnow")
}

func commandShutdown() {

	if os.Getuid() == 0 {
		// if the program is run by root user we are doing the most powerfull shutdown - that always shuts down the computer
		runCommand("shutdown", "-h", "now")
	} else {
		// if the program is run by ordinary user we are trying to shutdown, but it may fail if the other user is logged in
		runCommand("/usr/bin/osascript", "-e", "tell app \"System Events\" to shut down")
	}

}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

func publishDiscoveryMessages(client mqtt.Client) {
	prefix := getTopicPrefix()

	// Device information shared across all entities
	device := map[string]interface{}{
		"identifiers":  []string{"mac2mqtt_" + hostname},
		"name":         hostname,
		"model":        "macOS Computer",
		"manufacturer": "Apple",
		"sw_version":   "mac2mqtt",
	}

	// Binary sensor for alive status
	aliveConfig := map[string]interface{}{
		"name":         hostname + " Status",
		"unique_id":    "mac2mqtt_" + hostname + "_alive",
		"state_topic":  prefix + "/status/alive",
		"payload_on":   "true",
		"payload_off":  "false",
		"device_class": "connectivity",
		"device":       device,
	}
	publishConfig(client, "binary_sensor", hostname+"_alive", aliveConfig)

	// Sensor for battery
	batteryConfig := map[string]interface{}{
		"name":                  hostname + " Battery",
		"unique_id":             "mac2mqtt_" + hostname + "_battery",
		"state_topic":           prefix + "/status/battery",
		"unit_of_measurement":   "%",
		"device_class":          "battery",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "sensor", hostname+"_battery", batteryConfig)

	// Sensor for volume (read-only)
	volumeSensorConfig := map[string]interface{}{
		"name":                  hostname + " Volume Level",
		"unique_id":             "mac2mqtt_" + hostname + "_volume_sensor",
		"state_topic":           prefix + "/status/volume",
		"unit_of_measurement":   "%",
		"icon":                  "mdi:volume-high",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "sensor", hostname+"_volume_sensor", volumeSensorConfig)

	// Switch for mute
	muteConfig := map[string]interface{}{
		"name":                  hostname + " Mute",
		"unique_id":             "mac2mqtt_" + hostname + "_mute",
		"state_topic":           prefix + "/status/mute",
		"command_topic":         prefix + "/command/mute",
		"payload_on":            "true",
		"payload_off":           "false",
		"icon":                  "mdi:volume-mute",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "switch", hostname+"_mute", muteConfig)

	// Number for volume control
	volumeConfig := map[string]interface{}{
		"name":                  hostname + " Volume",
		"unique_id":             "mac2mqtt_" + hostname + "_volume",
		"state_topic":           prefix + "/status/volume",
		"command_topic":         prefix + "/command/volume",
		"min":                   0,
		"max":                   100,
		"step":                  1,
		"icon":                  "mdi:volume-medium",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "number", hostname+"_volume", volumeConfig)

	// Button for sleep
	sleepConfig := map[string]interface{}{
		"name":                  hostname + " Sleep",
		"unique_id":             "mac2mqtt_" + hostname + "_sleep",
		"command_topic":         prefix + "/command/sleep",
		"payload_press":         "sleep",
		"icon":                  "mdi:power-sleep",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "button", hostname+"_sleep", sleepConfig)

	// Button for shutdown
	shutdownConfig := map[string]interface{}{
		"name":                  hostname + " Shutdown",
		"unique_id":             "mac2mqtt_" + hostname + "_shutdown",
		"command_topic":         prefix + "/command/shutdown",
		"payload_press":         "shutdown",
		"icon":                  "mdi:power",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "button", hostname+"_shutdown", shutdownConfig)

	// Button for display sleep
	displaysleepConfig := map[string]interface{}{
		"name":                  hostname + " Display Sleep",
		"unique_id":             "mac2mqtt_" + hostname + "_displaysleep",
		"command_topic":         prefix + "/command/displaysleep",
		"payload_press":         "displaysleep",
		"icon":                  "mdi:monitor-off",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "button", hostname+"_displaysleep", displaysleepConfig)

	// Sensor for active application
	activeAppConfig := map[string]interface{}{
		"name":                  hostname + " Active App",
		"unique_id":             "mac2mqtt_" + hostname + "_active_app",
		"state_topic":           prefix + "/status/active_app",
		"icon":                  "mdi:application",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "sensor", hostname+"_active_app", activeAppConfig)

	// Sensor for Wi-Fi SSID
	wifiSSIDConfig := map[string]interface{}{
		"name":                  hostname + " Wi-Fi SSID",
		"unique_id":             "mac2mqtt_" + hostname + "_wifi_ssid",
		"state_topic":           prefix + "/status/wifi_ssid",
		"icon":                  "mdi:wifi",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "sensor", hostname+"_wifi_ssid", wifiSSIDConfig)

	// Sensor for Wi-Fi Signal Strength (RSSI)
	wifiSignalConfig := map[string]interface{}{
		"name":                  hostname + " Wi-Fi Signal Strength",
		"unique_id":             "mac2mqtt_" + hostname + "_wifi_signal_strength",
		"state_topic":           prefix + "/status/wifi_signal_strength",
		"unit_of_measurement":   "dBm",
		"icon":                  "mdi:wifi-strength-2",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "sensor", hostname+"_wifi_signal_strength", wifiSignalConfig)

	// Sensor for Wi-Fi IP Address
	wifiIPConfig := map[string]interface{}{
		"name":                  hostname + " Wi-Fi IP",
		"unique_id":             "mac2mqtt_" + hostname + "_wifi_ip",
		"state_topic":           prefix + "/status/wifi_ip",
		"icon":                  "mdi:ip-network",
		"availability_topic":    prefix + "/status/alive",
		"payload_available":     "true",
		"payload_not_available": "false",
		"device":                device,
	}
	publishConfig(client, "sensor", hostname+"_wifi_ip", wifiIPConfig)

	log.Println("Published Home Assistant MQTT discovery messages")
}

func publishConfig(client mqtt.Client, component, objectId string, config map[string]interface{}) {
	topic := fmt.Sprintf("homeassistant/%s/mac2mqtt_%s/%s/config", component, hostname, objectId)
	payload, err := json.Marshal(config)
	if err != nil {
		log.Printf("Error marshaling config for %s: %v", objectId, err)
		return
	}

	token := publishMQTT(client, topic, 0, true, payload)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Error publishing discovery for %s: %v", objectId, token.Error())
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected to MQTT")

	token := publishMQTT(client, getTopicPrefix()+"/status/alive", 0, true, "true")
	token.Wait()

	log.Println("Sending 'true' to topic: " + getTopicPrefix() + "/status/alive")

	// Publish Home Assistant discovery messages
	publishDiscoveryMessages(client)

	// Publish initial metrics
	updateVolume(client)
	updateMute(client)
	updateBattery(client)
	updateActiveApp(client)
	updateWiFiSSID(client)
	updateWiFiSignalStrength(client)
	updateWiFiIPAddress(client)

	listen(client, getTopicPrefix()+"/command/#")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Disconnected from MQTT: %v", err)
}

func getMQTTClient(ip, port, user, password string) mqtt.Client {
	// In dry-run mode, skip actual MQTT connection
	if dryRunMode {
		log.Println("Dry-run mode: Simulating MQTT connection")
		client := &dummyClient{}
		// Manually trigger the connect handler to simulate connection
		connectHandler(client)
		return client
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", ip, port))
	opts.SetUsername(user)
	opts.SetPassword(password)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	opts.SetWill(getTopicPrefix()+"/status/alive", "false", 0, true)

	// Enable automatic reconnection
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(60 * time.Second)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetConnectRetry(true)

	client := mqtt.NewClient(opts)

	// Retry initial connection with exponential backoff
	maxRetries := 10
	retryDelay := 2 * time.Second
	maxRetryDelay := 60 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to MQTT broker at %s:%s (attempt %d/%d)", ip, port, i+1, maxRetries)

		token := client.Connect()
		token.Wait()

		if token.Error() == nil {
			log.Println("Successfully connected to MQTT broker")
			return client
		}

		log.Printf("Failed to connect to MQTT: %v", token.Error())

		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)

			// Exponential backoff
			retryDelay *= 2
			if retryDelay > maxRetryDelay {
				retryDelay = maxRetryDelay
			}
		}
	}

	// If all retries failed, keep trying indefinitely with max delay
	log.Printf("Initial connection attempts failed. Will keep trying every %v...", maxRetryDelay)
	for {
		token := client.Connect()
		token.Wait()

		if token.Error() == nil {
			log.Println("Successfully connected to MQTT broker")
			return client
		}

		log.Printf("Failed to connect to MQTT: %v. Retrying in %v...", token.Error(), maxRetryDelay)
		time.Sleep(maxRetryDelay)
	}
}

func getTopicPrefix() string {
	return "mac2mqtt/" + hostname
}

// publishMQTT publishes a message to MQTT with optional debug logging
func publishMQTT(client mqtt.Client, topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	if dryRunMode || debugMode {
		prefix := "[DEBUG]"
		if dryRunMode {
			prefix = "[DRY-RUN]"
		}

		// Convert byte arrays to strings for better readability
		displayPayload := payload
		if bytePayload, ok := payload.([]byte); ok {
			displayPayload = string(bytePayload)
		}

		log.Printf("%s Publishing to topic '%s': %v (QoS=%d, Retained=%v)", prefix, topic, displayPayload, qos, retained)
	}

	if dryRunMode {
		// Return a dummy token that does nothing
		return &dummyToken{}
	}

	token := client.Publish(topic, qos, retained, payload)
	return token
}

// dummyToken is a no-op token for dry-run mode
type dummyToken struct{}

func (t *dummyToken) Wait() bool                     { return true }
func (t *dummyToken) WaitTimeout(time.Duration) bool { return true }
func (t *dummyToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
func (t *dummyToken) Error() error { return nil }

// dummyClient is a no-op MQTT client for dry-run mode
type dummyClient struct{}

func (c *dummyClient) IsConnected() bool       { return true }
func (c *dummyClient) IsConnectionOpen() bool  { return true }
func (c *dummyClient) Connect() mqtt.Token     { return &dummyToken{} }
func (c *dummyClient) Disconnect(quiesce uint) {}
func (c *dummyClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	return &dummyToken{}
}
func (c *dummyClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	return &dummyToken{}
}
func (c *dummyClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	return &dummyToken{}
}
func (c *dummyClient) Unsubscribe(topics ...string) mqtt.Token             { return &dummyToken{} }
func (c *dummyClient) AddRoute(topic string, callback mqtt.MessageHandler) {}
func (c *dummyClient) OptionsReader() mqtt.ClientOptionsReader {
	return mqtt.ClientOptionsReader{}
}

func listen(client mqtt.Client, topic string) {

	token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {

		if msg.Topic() == getTopicPrefix()+"/command/volume" {

			i, err := strconv.Atoi(string(msg.Payload()))
			if err == nil && i >= 0 && i <= 100 {

				setVolume(i)

				updateVolume(client)
				updateMute(client)

			} else {
				log.Println("Incorrect value")
			}

		}

		if msg.Topic() == getTopicPrefix()+"/command/mute" {

			b, err := strconv.ParseBool(string(msg.Payload()))
			if err == nil {
				setMute(b)

				updateVolume(client)
				updateMute(client)

			} else {
				log.Println("Incorrect value")
			}

		}

		if msg.Topic() == getTopicPrefix()+"/command/sleep" {

			if string(msg.Payload()) == "sleep" {
				commandSleep()
			}

		}

		if msg.Topic() == getTopicPrefix()+"/command/displaysleep" {

			if string(msg.Payload()) == "displaysleep" {
				commandDisplaySleep()
			}

		}

		if msg.Topic() == getTopicPrefix()+"/command/shutdown" {

			if string(msg.Payload()) == "shutdown" {
				commandShutdown()
			}

		}

	})

	token.Wait()
	if token.Error() != nil {
		log.Printf("Token error: %s\n", token.Error())
	}
}

func updateVolume(client mqtt.Client) {
	token := publishMQTT(client, getTopicPrefix()+"/status/volume", 0, false, strconv.Itoa(getCurrentVolume()))
	token.Wait()
}

func updateMute(client mqtt.Client) {
	token := publishMQTT(client, getTopicPrefix()+"/status/mute", 0, false, strconv.FormatBool(getMuteStatus()))
	token.Wait()
}

func getBatteryChargePercent() string {

	output := getCommandOutput("/usr/bin/pmset", "-g", "batt")

	// $ /usr/bin/pmset -g batt
	// Now drawing from 'Battery Power'
	//  -InternalBattery-0 (id=4653155)        100%; discharging; 20:00 remaining present: true

	r := regexp.MustCompile(`(\d+)%`)
	percent := r.FindStringSubmatch(output)[1]

	return percent
}

func updateBattery(client mqtt.Client) {
	token := publishMQTT(client, getTopicPrefix()+"/status/battery", 0, false, getBatteryChargePercent())
	token.Wait()
}

func getActiveApp() string {
	output := getCommandOutput("/usr/bin/osascript", "-e", "tell application \"System Events\" to get name of first application process whose frontmost is true")
	return output
}

func updateActiveApp(client mqtt.Client) {
	token := publishMQTT(client, getTopicPrefix()+"/status/active_app", 0, false, getActiveApp())
	token.Wait()
}

func getWiFiSSID() string {
	// Prefer networksetup (works even when airport binary is missing on newer macOS)
	if ssid, ok := getSSIDFromNetworksetup(); ok {
		return ssid
	}

	// Try ipconfig getsummary which can expose SSID on some macOS versions
	if ssid, ok := getSSIDFromIpconfig(); ok {
		return ssid
	}

	// Try CoreWLAN via swift as a robust fallback
	if ssid, _, ok := getWiFiInfoViaSwift(); ok && ssid != "" {
		return ssid
	}

	output := getAirportInfo()

	// Extract SSID from airport output
	r := regexp.MustCompile(`\s+SSID: (.+)`)
	matches := r.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}

	return "Not Connected"
}

func getWiFiSignalStrength() string {
	output := getAirportInfo()

	// Extract RSSI (signal strength) from airport output
	r := regexp.MustCompile(`\s+agrCtlRSSI: (-?\d+)`)
	matches := r.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback to system_profiler (RSSI is present even when SSID is redacted)
	if rssi, ok := getRSSIFromSystemProfiler(); ok {
		return rssi
	}

	// Try CoreWLAN via swift
	if _, rssi, ok := getWiFiInfoViaSwift(); ok && rssi != "" {
		return rssi
	}

	return "0"
}

func getWiFiIPAddress() string {
	iface := getWiFiInterface()
	if iface == "" {
		return "Not Connected"
	}

	output := getCommandOutput("/usr/sbin/ipconfig", "getifaddr", iface)
	if output == "" {
		return "Not Connected"
	}
	return output
}

func updateWiFiSSID(client mqtt.Client) {
	token := publishMQTT(client, getTopicPrefix()+"/status/wifi_ssid", 0, false, getWiFiSSID())
	token.Wait()
}

func updateWiFiSignalStrength(client mqtt.Client) {
	token := publishMQTT(client, getTopicPrefix()+"/status/wifi_signal_strength", 0, false, getWiFiSignalStrength())
	token.Wait()
}

func updateWiFiIPAddress(client mqtt.Client) {
	token := publishMQTT(client, getTopicPrefix()+"/status/wifi_ip", 0, false, getWiFiIPAddress())
	token.Wait()
}

func getAirportInfo() string {
	path := findAirportPath()
	if path == "" {
		return ""
	}

	cmd := exec.Command(path, "-I")
	stdout, err := cmd.Output()
	if err == nil {
		return strings.TrimSuffix(string(stdout), "\n")
	}

	log.Printf("Warning: failed to run %s: %v", path, err)
	return ""
}

func getSSIDFromNetworksetup() (string, bool) {
	for _, iface := range wifiInterfaceCandidates() {
		cmd := exec.Command("/usr/sbin/networksetup", "-getairportnetwork", iface)
		stdout, err := cmd.CombinedOutput()
		if err != nil && debugMode {
			log.Printf("Warning: networksetup -getairportnetwork %s failed: %v", iface, err)
		}

		network := regexp.MustCompile(`Current Wi-Fi Network: (.+)`).FindStringSubmatch(string(stdout))
		if len(network) > 1 {
			return network[1], true
		}
		// Skip if explicitly reports not associated
		if strings.Contains(string(stdout), "not associated") {
			continue
		}
	}
	return "", false
}

func getRSSIFromSystemProfiler() (string, bool) {
	cmd := exec.Command("/usr/sbin/system_profiler", "-detailLevel", "mini", "SPAirPortDataType")
	stdout, err := cmd.Output()
	if err != nil {
		if debugMode {
			log.Printf("Warning: system_profiler SPAirPortDataType failed: %v", err)
		}
		return "", false
	}

	rssi := regexp.MustCompile(`\s+RSSI: (-?\d+)`).FindStringSubmatch(string(stdout))
	if len(rssi) > 1 {
		return rssi[1], true
	}
	return "", false
}

// getWiFiInfoViaSwift uses CoreWLAN via the Swift interpreter to fetch SSID and RSSI.
func getWiFiInfoViaSwift() (string, string, bool) {
	script := `
import CoreWLAN
if let iface = CWWiFiClient.shared().interface() {
    if let ssid = iface.ssid() {
        print("SSID:\(ssid)")
    }
    print("RSSI:\(iface.rssiValue())")
}
`
	cmd := exec.Command("/usr/bin/swift", "-e", script)

	cacheDir := getSwiftCacheDir()
	env := os.Environ()
	if cacheDir != "" {
		env = append(env, "SWIFT_MODULE_CACHE_PATH="+cacheDir)
		env = append(env, "CLANG_MODULE_CACHE_PATH="+cacheDir)
	}
	cmd.Env = env

	stdout, err := cmd.CombinedOutput()
	if err != nil {
		if debugMode {
			log.Printf("Warning: swift CoreWLAN SSID/RSSI failed: %v (%s)", err, strings.TrimSpace(string(stdout)))
		}
		return "", "", false
	}

	outStr := string(stdout)
	ssid := regexp.MustCompile(`(?m)^SSID:(.+)$`).FindStringSubmatch(outStr)
	rssi := regexp.MustCompile(`(?m)^RSSI:([-0-9]+)$`).FindStringSubmatch(outStr)

	ssidVal := ""
	rssiVal := ""
	if len(ssid) > 1 {
		ssidVal = strings.TrimSpace(ssid[1])
	}
	if len(rssi) > 1 {
		rssiVal = strings.TrimSpace(rssi[1])
	}

	if ssidVal == "" && rssiVal == "" {
		return "", "", false
	}
	return ssidVal, rssiVal, true
}

func getSSIDFromIpconfig() (string, bool) {
	for _, iface := range wifiInterfaceCandidates() {
		cmd := exec.Command("/usr/sbin/ipconfig", "getsummary", iface)
		stdout, err := cmd.CombinedOutput()
		if err != nil {
			if debugMode {
				log.Printf("Warning: ipconfig getsummary %s failed: %v", iface, err)
			}
			continue
		}

		ssid := regexp.MustCompile(`(?m)SSID:\\s*(.+)`).FindStringSubmatch(string(stdout))
		if len(ssid) > 1 {
			return strings.TrimSpace(ssid[1]), true
		}
	}
	return "", false
}

var swiftCacheDir string
var swiftCacheOnce sync.Once

func getSwiftCacheDir() string {
	swiftCacheOnce.Do(func() {
		dir, err := os.MkdirTemp("", "mac2mqtt-swiftcache")
		if err != nil {
			if debugMode {
				log.Printf("Warning: unable to create Swift cache dir: %v", err)
			}
			swiftCacheDir = ""
			return
		}
		swiftCacheDir = dir
	})
	return swiftCacheDir
}

// getWiFiInterface returns the device name (enX) of the Wi-Fi interface.
func getWiFiInterface() string {
	cmd := exec.Command("/usr/sbin/networksetup", "-listallhardwareports")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		if debugMode {
			log.Printf("Warning: failed to list hardware ports: %v", err)
		}
	}
	if len(stdout) == 0 {
		return ""
	}

	lines := strings.Split(string(stdout), "\n")
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "Hardware Port: Wi-Fi") || strings.HasPrefix(lines[i], "Hardware Port: AirPort") {
			// Next line should be "Device: enX"
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "Device: ") {
				return strings.TrimSpace(strings.TrimPrefix(lines[i+1], "Device: "))
			}
		}
	}

	// Fallback to en0 if detection fails
	return "en0"
}

// wifiInterfaceCandidates returns possible Wi-Fi interfaces to probe.
func wifiInterfaceCandidates() []string {
	seen := make(map[string]bool)
	var candidates []string

	add := func(iface string) {
		if iface == "" {
			return
		}
		if !seen[iface] {
			seen[iface] = true
			candidates = append(candidates, iface)
		}
	}

	// Preferred interface
	add(getWiFiInterface())

	// Common fallbacks
	add("en0")
	add("en1")
	add("en2")
	add("en3")

	return candidates
}

var airportPath string
var airportPathOnce sync.Once

func findAirportPath() string {
	airportPathOnce.Do(func() {
		candidates := []string{
			"/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport",
			"/System/Library/PrivateFrameworks/Apple80211.framework/Versions/A/Resources/airport",
		}

		// Add any other versioned airport binaries if present (e.g., B, C, etc.)
		if matches, err := filepath.Glob("/System/Library/PrivateFrameworks/Apple80211.framework/Versions/*/Resources/airport"); err == nil {
			candidates = append(candidates, matches...)
		}

		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				airportPath = p
				return
			}
		}
		airportPath = ""
	})

	return airportPath
}

func main() {

	log.Println("Started")

	var c config
	c.getConfig()

	var wg sync.WaitGroup

	hostname = getHostname()
	mqttClient := getMQTTClient(c.Ip, c.Port, c.User, c.Password)

	volumeTicker := time.NewTicker(2 * time.Second)
	batteryTicker := time.NewTicker(60 * time.Second)

	wg.Add(1)
	go func() {
		for {
			select {
			case _ = <-volumeTicker.C:
				updateVolume(mqttClient)
				updateMute(mqttClient)
				updateActiveApp(mqttClient)

			case _ = <-batteryTicker.C:
				updateBattery(mqttClient)
				updateWiFiSSID(mqttClient)
				updateWiFiSignalStrength(mqttClient)
				updateWiFiIPAddress(mqttClient)
			}
		}
	}()

	wg.Wait()

}
