package scalingo

import (
	"encoding/json"
	"fmt"
	"testing"
)

var eventsSpecializeCases = map[string]struct {
	Event *Event

	// Expected
	DetailedEventName   string
	DetailedEventString string
}{
	"test event specialization": {
		Event: &Event{
			User: EventUser{
				Username: "user1",
			},
			Type:        EventRestart,
			RawTypeData: json.RawMessage([]byte(`{"scope": ["web"]}`)),
		},
		DetailedEventName:   "*scalingo.EventRestartType",
		DetailedEventString: "containers [web] have been restarted",
	},
	"test edit app event without with null force_https": {
		Event: &Event{
			Type:        EventEditApp,
			RawTypeData: json.RawMessage([]byte(`{"force_https": null}`)),
		},
		DetailedEventName:   "*scalingo.EventEditAppType",
		DetailedEventString: "application settings have been updated",
	},
	"test edit app event, force https enabled": {
		Event: &Event{
			Type:        EventEditApp,
			RawTypeData: json.RawMessage([]byte(`{"force_https": true}`)),
		},
		DetailedEventName:   "*scalingo.EventEditAppType",
		DetailedEventString: "application settings have been updated, Force HTTPS has been enabled",
	},
	"test edit app event, force https disabled": {
		Event: &Event{
			Type:        EventEditApp,
			RawTypeData: json.RawMessage([]byte(`{"force_https": false}`)),
		},
		DetailedEventName:   "*scalingo.EventEditAppType",
		DetailedEventString: "application settings have been updated, Force HTTPS has been disabled",
	},
}

func TestEvent_Specialize(t *testing.T) {
	for msg, c := range eventsSpecializeCases {
		t.Run(msg, func(t *testing.T) {
			dev := c.Event.Specialize()

			tdev := fmt.Sprintf("%T", dev)
			if tdev != c.DetailedEventName {
				t.Errorf("Expecting event of type %v, got %v", c.DetailedEventName, tdev)
			}

			if dev.String() != c.DetailedEventString {
				t.Errorf("Expecting event string\n===\n%s\n=== got\n%s\n===", c.DetailedEventString, dev.String())
			}
		})
	}
}
