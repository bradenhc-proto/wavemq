package wavemq

import (
	"encoding/gob"
	"reflect"
)

// Client ...
type Client struct {
	Name        string
	Persist     bool
	Sessions    map[string]Session
	publishers  map[string]*PublishChannel
	subscribers map[string]*SubscribeChannel
	messages    map[string]bool
}

// Connect ... returns the session id, which can be used as the key to restore the session
func (c *Client) Connect(server string, properties ConnectProperties) (string, error) {
	return "", nil
}

// Reconnect ...
func (c *Client) Reconnect(sessionID string) error {
	return nil
}

// Close ...
func (c *Client) Close() error {
	return nil
}

// SubscribeTo ...
func (c *Client) SubscribeTo(topic Topic) *SubscribeChannel {
	c.registerMessage(topic.Message)
	return NewSubscribeChannel(&topic)
}

// PublishOn ...
func (c *Client) PublishOn(topic Topic) *PublishChannel {
	c.registerMessage(topic.Message)
	return NewPublishChannel(&topic)
}

// registerMessage will add the provided message type to the list of messages that this client knows how to
// process. Types can only be registered once. If the type is registered again, then this function will return false.
// Returns true if the message is successfully registered.
func (c *Client) registerMessage(message interface{}) bool {
	t := reflect.TypeOf(message).String()
	if !c.messages[t] {
		c.messages[t] = true
		_, encodeable := message.(Encodeable)
		if !encodeable {
			gob.Register(message)
		}
		return true
	}
	return false
}
