package newsman

import (
	"net/url"
	"reflect"
	"strings"
)

const (
	// All is all
	All = "*"

	// EventTypesURLKey is the key we look in urls for
	EventTypesURLKey = "event"

	// DataTypesURLKey is the key we look in urls for
	DataTypesURLKey = "type"

	// TopicsURLKey is the key we look in urls for
	TopicsURLKey = "topic"
)

var (
	allTypes = []string{All}
	// AllEventTypes is a reference type to let us know we need all notifications
	AllEventTypes = allTypes
	// AllDataTypes is a reference type to let us know we need all notifications
	AllDataTypes = allTypes
	// AllTopics is a reference type to let us know we need all notifications
	AllTopics = allTypes

	// AllInclusiveListenerConfig represents a listener that wants every notification
	AllInclusiveListenerConfig = &ListenerConfig{
		Events:    AllEventTypes,
		DataTypes: AllDataTypes,
		Topics:    AllTopics,
	}
)

// ListenerConfig determines what events we're notified about
type ListenerConfig struct {
	Events    []string
	DataTypes []string
	Topics    []string
}

// ParseConfigFromURL parses a feed query from an HTTP request's URL
func ParseConfigFromURL(q url.Values) *ListenerConfig {
	fq := &ListenerConfig{}

	if e, ok := q[EventTypesURLKey]; ok {
		fq.Events = e
	}

	if dt, ok := q[DataTypesURLKey]; ok {
		fq.DataTypes = dt
	}

	if t, ok := q[TopicsURLKey]; ok {
		fq.Topics = t
	}

	return fq
}

// Values is a helper function to build a url.Values from
// the contents of a ListenerConfig
func (fq *ListenerConfig) Values() url.Values {
	v := url.Values{}

	if fq.Events != nil {
		for _, x := range fq.Events {
			v.Add(EventTypesURLKey, x)
		}
	}
	if fq.DataTypes != nil {
		for _, x := range fq.DataTypes {
			v.Add(DataTypesURLKey, x)
		}
	}
	if fq.Topics != nil {
		for _, x := range fq.Topics {
			v.Add(TopicsURLKey, x)
		}
	}
	return v
}

func typeString(i interface{}) string {
	return strings.TrimPrefix(reflect.TypeOf(i).String(), "*")
}

// IsInterested determines whether or not a listener is interested in an event.
func (fq *ListenerConfig) IsInterested(event Event, typeNameManipulationFunc TypeNameManipulationFunc) bool {
	var (
		interestedInEvent    = len(fq.Events) == 1 && fq.Events[0] == All
		interestedInDataType = len(fq.DataTypes) == 1 && fq.DataTypes[0] == All
		interestedInTopic    = len(fq.Topics) == 1 && fq.Topics[0] == All
	)

	if !interestedInEvent {
		for _, et := range fq.Events {
			if event.EventType == et {
				interestedInEvent = true
				break
			}
		}
	}

	if !interestedInDataType {
		for _, dt := range fq.DataTypes {
			ts := typeString(event.Data)
			if typeNameManipulationFunc != nil {
				ts = typeNameManipulationFunc(ts)
			}

			if strings.ToLower(ts) == strings.ToLower(dt) {
				interestedInDataType = true
				break
			}
		}
	}

	if !interestedInTopic {
		for _, et := range fq.Topics {
			for _, t := range event.Topics {
				if et == t {
					interestedInTopic = true
					break
				}
			}
		}
	}

	return interestedInEvent && interestedInDataType && interestedInTopic
}
