package notification

import (
	"fmt"
	"sort"
	"strings"
	"tastinginvite/internal/model"
	"time"
)

type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelPrint Channel = "print"
)

type Message struct {
	ID        string
	RecordID  string
	Recipient string
	Channel   Channel
	Subject   string
	Body      string
	CreatedAt time.Time
	SentAt    *time.Time
	Attempts  int
	Status    string
}

type Template struct {
	Name     string
	Subject  string
	Body     string
	Required []string
}

type Dispatcher struct {
	queue []Message
	sent  []Message
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{queue: make([]Message, 0), sent: make([]Message, 0)}
}

func (d *Dispatcher) Enqueue(record model.Record, recipient string, channel Channel, template Template, at time.Time) (Message, error) {
	if record.ID == "" || recipient == "" {
		return Message{}, fmt.Errorf("message context missing")
	}
	if !ValidChannel(channel) {
		return Message{}, fmt.Errorf("unsupported channel")
	}
	body, err := Render(template, record)
	if err != nil {
		return Message{}, err
	}
	message := Message{ID: fmt.Sprintf("msg-%s-%d", record.ID, len(d.queue)+len(d.sent)+1), RecordID: record.ID, Recipient: recipient, Channel: channel, Subject: template.Subject, Body: body, CreatedAt: at, Status: "queued"}
	d.queue = append(d.queue, message)
	return message, nil
}

func (d *Dispatcher) Dispatch(now time.Time) []Message {
	delivered := make([]Message, 0, len(d.queue))
	remaining := make([]Message, 0)
	for _, message := range d.queue {
		message.Attempts++
		if strings.TrimSpace(message.Recipient) == "" {
			message.Status = "failed"
			remaining = append(remaining, message)
			continue
		}
		message.Status = "sent"
		message.SentAt = &now
		d.sent = append(d.sent, message)
		delivered = append(delivered, message)
	}
	d.queue = remaining
	return delivered
}

func (d *Dispatcher) RetryFailed() int {
	moved := 0
	remaining := make([]Message, 0, len(d.queue))
	for _, message := range d.queue {
		if message.Status == "failed" && message.Attempts < 3 {
			message.Status = "queued"
			remaining = append(remaining, message)
			moved++
		} else {
			remaining = append(remaining, message)
		}
	}
	d.queue = remaining
	return moved
}

func (d *Dispatcher) Pending() []Message {
	out := append([]Message(nil), d.queue...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (d *Dispatcher) Sent() []Message {
	out := append([]Message(nil), d.sent...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func Render(template Template, record model.Record) (string, error) {
	body := template.Body
	replacements := map[string]string{"{{title}}": record.Title, "{{host}}": record.Host, "{{venue}}": record.Venue, "{{status}}": record.Status, "{{capacity}}": fmt.Sprintf("%d", record.Capacity)}
	for _, required := range template.Required {
		if !strings.Contains(body, required) {
			return "", fmt.Errorf("template missing %s", required)
		}
	}
	for key, value := range replacements {
		body = strings.ReplaceAll(body, key, value)
	}
	return body, nil
}

func ValidChannel(channel Channel) bool {
	return channel == ChannelEmail || channel == ChannelSMS || channel == ChannelPrint
}

func Templates() map[string]Template {
	return map[string]Template{"invite": {Name: "invite", Subject: "Invitation: {{title}}", Body: "Join {{title}} at {{venue}} with {{host}}.", Required: []string{"{{title}}", "{{venue}}"}}, "reminder": {Name: "reminder", Subject: "Reminder: {{title}}", Body: "Your tasting at {{venue}} is coming up.", Required: []string{"{{title}}", "{{venue}}"}}}
}

func Digest(messages []Message) map[string]int {
	result := map[string]int{}
	for _, message := range messages {
		result[string(message.Channel)]++
	}
	return result
}
