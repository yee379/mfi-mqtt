package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	// "log"
	// "crypto/tls"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type PortStats struct {
	Port        int     `json:"port"`
	Output      int     `json:"output"`
	Power       float32 `json:"power"`
	Enabled     int     `json:"enabled"`
	Current     float32 `json:"current"`
	Voltage     float32 `json:"voltage"`
	Powerfactor float32 `json:"powerfactor"`
	Relay       int     `json:"relay"`
	Lock        int     `json:"lock"`
	// prevmonth uint8
	// thismonth uint8
}

type PortResponse struct {
	Sensors []PortStats `json:"sensors"`
	Status  string      `json:"status"`
}

const proc_root = "/proc/"

var file_mappings = map[string][]string{
	"Output":      []string{"power/output", "int"},
	"Power":       []string{"power/active_pwr", "float32"},
	"Enabled":     []string{"power/enabled", "int"},
	"Current":     []string{"power/i_rms", "float32"},
	"Voltage":     []string{"power/v_rms", "float32"},
	"Powerfactor": []string{"power/pf", "float32"},
	"Relay":       []string{"power/relay", "int"},
	"Lock":        []string{"power/lock", "int"},
}

func get_port_stats(port int) PortStats {
	stats := PortStats{Port: port}
	for key, value := range file_mappings {
		filename := value[0]
		content_type := value[1]
		filepath := filepath.Join(proc_root, filename+strconv.Itoa(port))
		content, err := ioutil.ReadFile(filepath)
		if err != nil {
			fmt.Errorf("failed to open procfs: %v", err)
		}
		value := strings.TrimSpace(string(content))
		if content_type == "int" {
			v, _ := strconv.ParseInt(value, 10, 8)
			reflect.ValueOf(&stats).Elem().FieldByName(key).SetInt(v)
		} else if content_type == "float32" {
			v, _ := strconv.ParseFloat(value, 32)
			reflect.ValueOf(&stats).Elem().FieldByName(key).SetFloat(v)
		}
	}
	return stats
}

func main() {

	hostname, _ := os.Hostname()

	server := flag.String("server", "tcp://127.0.0.1:1883", "The full URL of the MQTT server to connect to")
	topic := flag.String("topic", hostname, "Base topic to publish the messages on")
	qos := flag.Int("qos", 0, "The QoS to send the messages at")
	retained := flag.Bool("retained", false, "Are the messages sent with the retained flag")
	verbose := flag.Bool("verbose", false, "Spit out debugging messages")
	clientid := flag.String("clientid", hostname+strconv.Itoa(time.Now().Second()), "A clientid for the connection")
	username := flag.String("username", "", "A username to authenticate to the MQTT server")
	password := flag.String("password", "", "Password to match username")
	// period := flag.Uint64("period", 30, "Frequency of sensor updates")
	flag.Parse()

	connOpts := MQTT.NewClientOptions().AddBroker(*server).SetClientID(*clientid).SetCleanSession(true)
	if *username != "" {
		connOpts.SetUsername(*username)
		if *password != "" {
			connOpts.SetPassword(*password)
		}
	}
	// tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	// connOpts.SetTLSConfig(tlsConfig)

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if *verbose {
		fmt.Println("Connected to %s\n", *server)
	}

	files, _ := filepath.Glob("/proc/power/enabled*")
	num_ports := len(files)

	for x := 1; x <= num_ports; x++ {
		stats := get_port_stats(x)
		json_data, err := json.Marshal(stats)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		this_topic := *topic + "/port" + strconv.Itoa(x)
		if *verbose {
			fmt.Println("sending ", this_topic, string(json_data))
		}
		token := client.Publish(this_topic, byte(*qos), *retained, json_data)
		token.Wait()
	}

	os.Exit(0)

}
