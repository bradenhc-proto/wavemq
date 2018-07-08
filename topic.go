package wavemq

// Topic represents a publish/subscribe topic in the MQTT protocol. A topic essentially consists of name (string) and
// a message to send (interface, since it could be anything). Topics also keep track of their encoder and decoder.
type Topic struct {
	Name    string
	Message interface{}
}
