package redeye

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Command is the JSON payload sent to the camera command topic.
//
// Supported commands:
//
//	snap   Save the current frame. Optional "file" sets the output path.
type Command struct {
	Command string `json:"command"`
	File    string `json:"file,omitempty"`
}

// MessengerStatus is returned by Status() and the /api/messenger endpoint.
type MessengerStatus struct {
	Broker        string   `json:"broker"`
	Connected     bool     `json:"connected"`
	Subscriptions []string `json:"subscriptions"`
}

// Messenger handles MQTT pub/sub for a single camera node. Connect must be
// called before Announce or SubscribeCommands.
type Messenger struct {
	Name   string // MQTT client ID, defaults to os.Hostname
	Broker string // broker URL, e.g. "tcp://localhost:1883"
	Prefix string // topic namespace, e.g. "/redeye"

	mu            sync.Mutex
	subscriptions []string
	client        mqtt.Client
	quit          chan struct{}
	wg            sync.WaitGroup
}

// NewMessenger creates a Messenger using broker and prefix. The MQTT client ID
// is set to the machine hostname; it falls back to "redeye" on error.
func NewMessenger(broker, prefix string) *Messenger {
	name, err := os.Hostname()
	if err != nil || name == "" {
		name = "redeye"
	}
	return &Messenger{
		Name:   name,
		Broker: broker,
		Prefix: prefix,
		quit:   make(chan struct{}),
	}
}

// Connect establishes the MQTT broker connection. Returns an error if the
// broker is unreachable; the Messenger is safe to Close even on error.
func (m *Messenger) Connect() error {
	opts := mqtt.NewClientOptions().
		AddBroker(m.Broker).
		SetClientID(m.Name).
		SetKeepAlive(30 * time.Second).
		SetDefaultPublishHandler(m.handleMessage)

	m.client = mqtt.NewClient(opts)
	tok := m.client.Connect()
	if tok.Wait() && tok.Error() != nil {
		m.client = nil // clear partial paho state
		return fmt.Errorf("MQTT: connect to %s: %w", m.Broker, tok.Error())
	}
	log.Printf("MQTT: connected to %s as %q", m.Broker, m.Name)
	return nil
}

// Announce publishes this node's identity on <prefix>/announce/camera.
// Other nodes subscribed to that topic will discover this camera.
func (m *Messenger) Announce() error {
	payload, _ := json.Marshal(map[string]string{"id": m.Name})
	return m.publish(m.AnnounceTopic(), string(payload))
}

// SubscribeCommands subscribes to <prefix>/camera/<id> so this node receives
// JSON Command messages addressed to it.
func (m *Messenger) SubscribeCommands() error {
	return m.subscribe(m.CommandTopic())
}

// Status returns a snapshot of the current broker connection and subscriptions.
func (m *Messenger) Status() MessengerStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return MessengerStatus{
		Broker:        m.Broker,
		Connected:     m.client != nil && m.client.IsConnected(),
		Subscriptions: append([]string(nil), m.subscriptions...),
	}
}

// Close disconnects from the broker and waits for in-flight goroutines.
// Safe to call more than once.
func (m *Messenger) Close() {
	select {
	case <-m.quit:
	default:
		close(m.quit)
	}
	m.wg.Wait()
	if m.client != nil && m.client.IsConnected() {
		m.client.Disconnect(500)
	}
}

// AnnounceTopic returns the topic this node publishes its presence on.
func (m *Messenger) AnnounceTopic() string {
	return m.Prefix + "/announce/camera"
}

// CommandTopic returns the topic this node subscribes to for commands.
func (m *Messenger) CommandTopic() string {
	return m.Prefix + "/camera/" + m.Name
}

func (m *Messenger) subscribe(topic string) error {
	tok := m.client.Subscribe(topic, 0, nil)
	if tok.Wait() && tok.Error() != nil {
		return fmt.Errorf("MQTT: subscribe to %s: %w", topic, tok.Error())
	}
	m.mu.Lock()
	m.subscriptions = append(m.subscriptions, topic)
	m.mu.Unlock()
	log.Printf("MQTT: subscribed to %s", topic)
	return nil
}

func (m *Messenger) publish(topic, payload string) error {
	if m.client == nil || !m.client.IsConnected() {
		return fmt.Errorf("MQTT: not connected")
	}
	tok := m.client.Publish(topic, 0, false, payload)
	tok.Wait()
	if tok.Error() != nil {
		return fmt.Errorf("MQTT: publish to %s: %w", topic, tok.Error())
	}
	return nil
}

// handleMessage is called by the paho library for every incoming message on
// any subscribed topic. Commands are dispatched to the active Snapper.
func (m *Messenger) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	var cmd Command
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		log.Printf("MQTT: malformed payload on %s: %v", msg.Topic(), err)
		return
	}
	log.Printf("MQTT [in] topic=%s command=%q", msg.Topic(), cmd.Command)

	switch cmd.Command {
	case "snap":
		file := cmd.File
		if file == "" {
			file = fmt.Sprintf("snapshot-%s.jpg", time.Now().Format("20060102-150405"))
		}
		if s := getSnapper(); s != nil {
			if err := s.Snap(file); err != nil {
				log.Printf("MQTT snap error: %v", err)
			} else {
				log.Printf("MQTT snap: saved %s", file)
			}
		} else {
			log.Println("MQTT snap: no active camera source")
		}
	default:
		log.Printf("MQTT: unknown command %q on %s", cmd.Command, msg.Topic())
	}
}
