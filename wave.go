package wavemq

// ConnectionProperties ...
type ConnectionProperties struct {
	Username         string
	Password         []byte
	WillTopic        string
	WillMessage      string
	KeepAliveTimeout uint16
}

// Session ...
type Session struct {
	Name          string
	ServerAddress string
	identifier    string
	Flags         Flags
	state         interface{}
}

// Topic represents the identifier for the topic and the structure of the message that will
// be discussed. The structure is used to register the type with the internal encoder.
type Topic struct {
	name    string
	message interface{}
}

// Client ...
type Client struct {
	Persist  bool
	Sessions map[string]Session
}

// Connect ... returns the session name
func (c *Client) Connect(server string, properties ConnectionProperties) (string, error) {
	return "", nil
}

// Reconnect ...
func (c *Client) Reconnect(session Session) error {
	return nil
}

// Close ...
func (c *Client) Close() error {
	return nil
}

// SubscribeTo ...
func (c *Client) SubscribeTo(topic Topic) SubscribeChannel {
	return SubscribeChannel{}
}

// PublishOn ...
func (c *Client) PublishOn(topic Topic) PublishChannel {
	return PublishChannel{}
}

// SubscribeChannel ...
type SubscribeChannel struct {
}

// Receive ...
func (sc *SubscribeChannel) Receive() []byte {
	return make([]byte, 1)
}

// ReceiveIn ...
func (sc *SubscribeChannel) ReceiveIn(target interface{}) {

}

// PublishChannel ...
type PublishChannel struct {
}

// Send ...
func (pc *PublishChannel) Send(message interface{}) {

}
