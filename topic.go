package wavemq

// Topic represents a publish/subscribe topic in the MQTT protocol. A topic essentially consists of name (string) and
// a message to send (interface, since it could be anything). Topics also keep track of their encoder and decoder.
// Topics are uniquely identifiable by the combination of the name and the message, so two topics that have the same
// name but are handling separate messages would be two completely separate topics in WaveMQ.
type Topic struct {
	Name    string
	Message interface{}
}
