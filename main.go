package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var hostname string
var debugMode bool

type config struct {
	Ip       string `yaml:"mqtt_ip"`
	Port     string `yaml:"mqtt_port"`
	User     string `yaml:"mqtt_user"`
	Password string `yaml:"mqtt_password"`
	Debug    bool   `yaml:"debug"`
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

	if c.Ip == "" {
		log.Fatal("Must specify mqtt_ip in mac2mqtt.yaml")
	}

	if c.Port == "" {
		log.Fatal("Must specify mqtt_port in mac2mqtt.yaml")
	}

	// Set global debug mode
	debugMode = c.Debug

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
		"name":              hostname + " Status",
		"unique_id":         "mac2mqtt_" + hostname + "_alive",
		"state_topic":       prefix + "/status/alive",
		"payload_on":        "true",
		"payload_off":       "false",
		"device_class":      "connectivity",
		"device":            device,
	}
	publishConfig(client, "binary_sensor", hostname+"_alive", aliveConfig)

	// Sensor for battery
	batteryConfig := map[string]interface{}{
		"name":              hostname + " Battery",
		"unique_id":         "mac2mqtt_" + hostname + "_battery",
		"state_topic":       prefix + "/status/battery",
		"unit_of_measurement": "%",
		"device_class":      "battery",
		"availability_topic": prefix + "/status/alive",
		"payload_available":  "true",
		"payload_not_available": "false",
		"device":            device,
	}
	publishConfig(client, "sensor", hostname+"_battery", batteryConfig)

	// Sensor for volume (read-only)
	volumeSensorConfig := map[string]interface{}{
		"name":              hostname + " Volume Level",
		"unique_id":         "mac2mqtt_" + hostname + "_volume_sensor",
		"state_topic":       prefix + "/status/volume",
		"unit_of_measurement": "%",
		"icon":              "mdi:volume-high",
		"availability_topic": prefix + "/status/alive",
		"payload_available":  "true",
		"payload_not_available": "false",
		"device":            device,
	}
	publishConfig(client, "sensor", hostname+"_volume_sensor", volumeSensorConfig)

	// Switch for mute
	muteConfig := map[string]interface{}{
		"name":              hostname + " Mute",
		"unique_id":         "mac2mqtt_" + hostname + "_mute",
		"state_topic":       prefix + "/status/mute",
		"command_topic":     prefix + "/command/mute",
		"payload_on":        "true",
		"payload_off":       "false",
		"icon":              "mdi:volume-mute",
		"availability_topic": prefix + "/status/alive",
		"payload_available":  "true",
		"payload_not_available": "false",
		"device":            device,
	}
	publishConfig(client, "switch", hostname+"_mute", muteConfig)

	// Number for volume control
	volumeConfig := map[string]interface{}{
		"name":              hostname + " Volume",
		"unique_id":         "mac2mqtt_" + hostname + "_volume",
		"state_topic":       prefix + "/status/volume",
		"command_topic":     prefix + "/command/volume",
		"min":               0,
		"max":               100,
		"step":              1,
		"icon":              "mdi:volume-medium",
		"availability_topic": prefix + "/status/alive",
		"payload_available":  "true",
		"payload_not_available": "false",
		"device":            device,
	}
	publishConfig(client, "number", hostname+"_volume", volumeConfig)

	// Button for sleep
	sleepConfig := map[string]interface{}{
		"name":              hostname + " Sleep",
		"unique_id":         "mac2mqtt_" + hostname + "_sleep",
		"command_topic":     prefix + "/command/sleep",
		"payload_press":     "sleep",
		"icon":              "mdi:power-sleep",
		"availability_topic": prefix + "/status/alive",
		"payload_available":  "true",
		"payload_not_available": "false",
		"device":            device,
	}
	publishConfig(client, "button", hostname+"_sleep", sleepConfig)

	// Button for shutdown
	shutdownConfig := map[string]interface{}{
		"name":              hostname + " Shutdown",
		"unique_id":         "mac2mqtt_" + hostname + "_shutdown",
		"command_topic":     prefix + "/command/shutdown",
		"payload_press":     "shutdown",
		"icon":              "mdi:power",
		"availability_topic": prefix + "/status/alive",
		"payload_available":  "true",
		"payload_not_available": "false",
		"device":            device,
	}
	publishConfig(client, "button", hostname+"_shutdown", shutdownConfig)

	// Button for display sleep
	displaysleepConfig := map[string]interface{}{
		"name":              hostname + " Display Sleep",
		"unique_id":         "mac2mqtt_" + hostname + "_displaysleep",
		"command_topic":     prefix + "/command/displaysleep",
		"payload_press":     "displaysleep",
		"icon":              "mdi:monitor-off",
		"availability_topic": prefix + "/status/alive",
		"payload_available":  "true",
		"payload_not_available": "false",
		"device":            device,
	}
	publishConfig(client, "button", hostname+"_displaysleep", displaysleepConfig)

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

	listen(client, getTopicPrefix()+"/command/#")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Disconnected from MQTT: %v", err)
}

func getMQTTClient(ip, port, user, password string) mqtt.Client {

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
	if debugMode {
		log.Printf("[DEBUG] Publishing to topic '%s': %v (QoS=%d, Retained=%v)", topic, payload, qos, retained)
	}
	token := client.Publish(topic, qos, retained, payload)
	return token
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

			case _ = <-batteryTicker.C:
				updateBattery(mqttClient)
			}
		}
	}()

	wg.Wait()

}
