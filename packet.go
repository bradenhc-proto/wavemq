package wavemq

import (
	"bytes"
	"encoding/gob"
	"errors"
)

// The following constants define the control packet types and their assocaited flags for the MQTT 3.1.1 protocol.
// The control type together with the flag represents the first byte if each packet sent in MQTT. In order to
// create this first byte, WaveMQ will perform a bitwise OR operation on the two bytes representing the type and
// flags respectively.
const (
	// ptypeConnect is the control packet type for client request to connect to a server
	ptypeConnect byte = 0x10
	// ptypeConnack is the control packet type for connection acknowledgment by the server
	ptypeConnack byte = 0x20
	// ptypePublish is the control packet type for a publish message sent from client to server or
	// server to client
	ptypePublish byte = 0x30
	// ptypePuback is the control packet for an acknowledged PUBLISH packet. It can be sent from client
	// to server or server to client.
	ptypePuback byte = 0x40
	// ptypePubrec is the publish received control packet type (part one of assured delivery). Sent by
	// the client or server.
	ptypePubrec byte = 0x50
	// ptypePubrel is the publish released control packet type (part two of assured delivery). Sent by
	// the client or server.
	ptypePubrel byte = 0x60
	// ptypePubcomp is the publish complete control packet type (part three of assured delivery). Sent
	// by client or server.
	ptypePubcomp byte = 0x70
	// ptypeSubscribe is the control packet for a client subscrbe request. Sent by the client.
	ptypeSubscribe byte = 0x80
	// ptypeSuback is the control packet type for a client subscription acknowledgment. Sent by the
	// the server.
	ptypeSuback byte = 0x90
	// ptypeUnsubscribe is the control packet type for a unsubscribe request sent by the client
	ptypeUnsubscribe byte = 0xA0
	// ptypeUnsuback is the control packet type for a unsebscribe acknowledgment sent by the server
	ptypeUnsuback byte = 0xB0
	// ptypePingreq is the control packet type for a ping request from client to server
	ptypePingreq byte = 0xC0
	// ptypePingresp is the control packet type for a ping response from server to client
	ptypePingresp byte = 0xD0
	// ptypeDisconnect is the control packet type for the client disconnecting from the server
	ptypeDisconnect byte = 0xE0
	// pflagsConnect represents the flags associated with the CONNECT control packet type
	pflagsConnect byte = 0x00
	// pflagsConnack represents the flags associated with the CONNACK control packet type
	pflagsConnack byte = 0x00
	// pflagsPuback represents the flags associated with the PUBACK control packet type
	pflagsPuback byte = 0x00
	// pflagsPubrec represents the flags associated with the PUBREC control packet type
	pflagsPubrec byte = 0x00
	// pflagsPubrel represents the flags associated with the PUBREL control packet type
	pflagsPubrel byte = 0x02
	// pflagsPubcomp represents the flags associated with the PUBCOMP control packet type
	pflagsPubcomp byte = 0x00
	// pflagsSubscribe represents the flags associated with the SUBSCRIBE control packet type
	pflagsSubscribe byte = 0x02
	// pflagsSuback represents the flags associated with the SUBACK control packet type
	pflagsSuback byte = 0x00
	// pflagsUnsubscribe represents the flags associated with the UNSUBSCRIBE control packet type
	pflagsUnsubscribe byte = 0x02
	// pflagsUnsuback represents the flags associated with the UNSUBACK control packet type
	pflagsUnsuback byte = 0x00
	// pflagsPingreq represents the flags associated with the PINGREQ control packet type
	pflagsPingreq byte = 0x00
	// pflagsPingresp represents the flags associated with the PINGRESP control packet type
	pflagsPingresp byte = 0x00
	// pflagsDisconnect represents the flags associated with the DISCONNECT control packet type
	pflagsDisconnect byte = 0x00
)

// QoSLevel represents a byte defining the level of quality of service. This is used to restrict the developer to
// only using the constants defined below as valid types
type QoSLevel byte

// The following constants define the quality of service flags used in the first byte of a PUBLISH packet. They can
// be bitwise OR'd with the first byte to flip the correct flag bits without any extra bitwise operations on part of
// the developer.
const (
	// QoSAtMostOnce is a Quality of Service that says a PUBLISH packet will be delivered at most once to the
	// subscriber
	QoSAtMostOnce QoSLevel = 0x00
	// QoSAtLeastOnce is a Quality of Service that says a PUBLISH packet will be delivered at least once to the
	// subscriber
	QoSAtLeastOnce QoSLevel = 0x02
	// QoSExactlyOnce is a Quality of Service that says a PUBLISH packet will be delievered exactly once to the
	// subscriber
	QoSExactlyOnce QoSLevel = 0x04
)

// packet represents a MQTT packet that is either sent or received.
type packet struct {
	ptype      byte
	pflags     byte
	length     uint32
	properties Encodeable
	payload    interface{}
	buffer     bytes.Buffer
}

// Encodeable defines an interface that, when implemented, specifies how a structure is to be encoded into a buffer
// (i.e. a slice of bytes). All variable headers (packet properties in WaveMQ) implement this interface. Some payloads
// sent as part of control packets other than PUBLISH also implement this interface, and it is up to WaveMQ to know
// whether a structure implements the interface or not.
//
// Payloads sent as part of a PUBLISH message can implement this interface if the user wishes to specify how to
// serialize the object. Without an implementation of the interface, WaveMQ will use the 'encode/gob' library to
// serialize and deserialize data structures
type Encodeable interface {
	Encode() ([]byte, error)
}

// ConnectProperties summarizes the properties found in the variable header of the CONNECT
// control type packet.
type ConnectProperties struct {
	ProtocolName  string
	ProtocolLevel int
	CleanSession  bool
	WillFlag      bool
	WillQoS       bool
	WillRetain    bool
	UserName      bool
	Password      bool
	KeepAlive     uint16
}

// Encode represents the implementation of the Encodeable interface and describes how the
// ConnectProperties variable header definition should be encoded in the packet
func (h ConnectProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	// Write the variable header
	protocolNameBytes := []byte(h.ProtocolName)
	buffer.WriteByte(byte(len(h.ProtocolName)))
	buffer.Write(protocolNameBytes)
	buffer.WriteByte(byte(h.ProtocolLevel))

	// Check the flags and write it to the buffer
	var flagsByte byte
	if h.CleanSession {
		flagsByte |= 0x02
	}
	if h.WillFlag {
		flagsByte |= 0x04
	}
	if h.WillQoS {
		flagsByte |= 0x08
	}
	if h.WillRetain {
		flagsByte |= 0x20
	}
	if h.Password {
		flagsByte |= 0x40
	}
	if h.UserName {
		flagsByte |= 0x80
	}
	buffer.WriteByte(flagsByte)

	// Write the keep alive time
	buffer.WriteByte(byte(h.KeepAlive & 0xF0))
	buffer.WriteByte(byte(h.KeepAlive & 0x0F))

	buf = buffer.Bytes()

	return buf, err
}

// ConnectPayload defines the attributes of the payload for a CONNECT control packet. These
// values will be encoded as length-prefixed fields
type ConnectPayload struct {
	Identifier  string
	WillTopic   string
	WillMessage string
	UserName    string
	Password    []byte
}

// Encode writes the payload content for a CONNECT control packet, which has a specific format. This is an
// implementation of the Encodeable interface.
func (p ConnectPayload) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}

	// Encode the identifier
	// TODO: verify the identifier is UTF-8 encoded and contains only valid characters
	bytes := []byte(p.Identifier)
	length := uint16(len(bytes))
	buffer.WriteByte(byte(length & 0xF0))
	buffer.WriteByte(byte(length & 0x0F))
	buffer.Write(bytes)

	// Encode the will topic
	// TODO: verify that the topic is UTF-8 encoded
	bytes = []byte(p.WillTopic)
	length = uint16(len(bytes))
	buffer.WriteByte(byte(length & 0xF0))
	buffer.WriteByte(byte(length & 0x0F))
	buffer.Write(bytes)

	// Encode the will message
	bytes = []byte(p.WillMessage)
	length = uint16(len(bytes))
	buffer.WriteByte(byte(length & 0xF0))
	buffer.WriteByte(byte(length & 0x0F))
	buffer.Write(bytes)

	// Encode the user name
	// TODO: verify the username is UTF-8 encoded
	bytes = []byte(p.UserName)
	length = uint16(len(bytes))
	buffer.WriteByte(byte(length & 0xF0))
	buffer.WriteByte(byte(length & 0x0F))
	buffer.Write(bytes)

	// Encode the password
	length = uint16(len(p.Password))
	buffer.WriteByte(byte(length & 0xF0))
	buffer.WriteByte(byte(length & 0x0F))
	buffer.Write(p.Password)

	return buf, err
}

// ConnectAckProperties summarizes the properties found in the variable header of the CONNECTACK
// control type packet
type ConnectAckProperties struct {
	SessionPresent bool
	ReturnCode     int
}

// Encode represents the implementation of the Encodeable interface and describes how the
// ConnectAckProperties variable header definition should be encoded in the packet
func (h ConnectAckProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	flags := byte(0x00)
	if h.SessionPresent {
		flags |= 0x01
	}
	buffer.WriteByte(flags)
	buffer.WriteByte(byte(h.ReturnCode))
	buf = buffer.Bytes()

	return buf, err
}

// PublishProperties summarizes the properties found in the variable header of the PUBLISH control type packet. It also
// includes the control packet flags since these can be set dynamically by the client/server (as oppsed to all the
// other packets who have fixed control type flags).
type PublishProperties struct {
	DupFlag   bool
	QoSLevel  QoSLevel
	Retain    bool
	TopicName string
	PacketID  uint16
}

// Encode writes the fields of the PublishProperties struct to a properly formated byte buffer that can be used as
// the variable header for a PUBLISH control packet.
func (h PublishProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	// encode the topic name
	topicNameBytes := []byte(h.TopicName)
	topicNameLength := uint16(len(topicNameBytes))
	buffer.WriteByte(byte(topicNameLength & 0xF0))
	buffer.WriteByte(byte(topicNameLength & 0x0F))

	// Encode the packet ID
	buffer.WriteByte(byte(h.PacketID & 0xF0))
	buffer.WriteByte(byte(h.PacketID & 0x0F))

	buf = buffer.Bytes()

	return buf, err
}

// PublishAckProperties defines the fields of the variable header for a PUBACK packet.
type PublishAckProperties struct {
	PacketID uint16
}

// Encode writes the variable header of a PUBACK message to byte buffer with fields as defined by the PubAckProperties
// struct. This is an implementation of the Encodeable interface.
func (h PublishAckProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	buffer.WriteByte(byte(h.PacketID & 0xF0))
	buffer.WriteByte(byte(h.PacketID & 0x0F))
	buf = buffer.Bytes()
	return buf, err
}

// PublishRecProperties defines the fields of the variable header for the PUBREC packet.
type PublishRecProperties struct {
	PacketID uint16
}

// Encode writes the variable header of a PUBREC message to a byte buffer using the fields and values from the
// PublishRecProperties struct. This is an implementation of the Encodeable interface.
func (h PublishRecProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	buffer.WriteByte(byte(h.PacketID & 0xF0))
	buffer.WriteByte(byte(h.PacketID & 0x0F))
	buf = buffer.Bytes()
	return buf, err
}

// PublishRelProperties defines the fields of the variable header for the PUBREL packet
type PublishRelProperties struct {
	PacketID uint16
}

// Encode writes the variable header of the PUBREL message to a byte buffer using the fields and values from the
// PublishRelProperties struct. This is an implementation of the Encodeable interface.
func (h PublishRelProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	buffer.WriteByte(byte(h.PacketID & 0xF0))
	buffer.WriteByte(byte(h.PacketID & 0x0F))
	buf = buffer.Bytes()
	return buf, err
}

// encodeRemainingLength operates on a pointer a packet struct by modifying its internal buffer
// to contain the provided length value in the encoded format specified in the MQTT protocol
// specifications. It will also update the internal offset of the packet so that the rest of
// the packet can be created. This function should only be called when building a packet to send
func encodeRemainingLength(p *packet, length uint32) {
	var encoded byte
	for length > 0 {
		encoded = byte(length % 0x80)
		length /= 0x80
		if length > 0 {
			encoded |= 0x80
		}
		p.buffer.WriteByte(encoded)
	}
}

// decodeRemainingLength looks at a packet's internal buffer and decodes the value of the
// remaining length s. It also updates the internal read offset by first setting it to 1 (the
// expected location of the start of the remaining length s) and incrementing it so that it
// ends at the start of the variable length header or payload (depending on the packet type)
func decodeRemainingLength(p *packet) (value uint32, err error) {
	var multiplier uint32 = 1
	var encoded byte
	first := true
	for (encoded&0x80) != 0 || first {
		if first {
			first = false
		}
		encoded, err = p.buffer.ReadByte()
		if err != nil {
			return 0, err
		}
		value += uint32(encoded&0x7F) * multiplier
		multiplier *= 128
		if multiplier > 128*128*128 {
			err = errors.New("Malformed remaining length")
			return 0, err
		}
	}
	return value, err
}

// encode writes the information in the packet to the internal buffer in preparation for
// delivery.
// TODO: when encoding, we need to make sure that byte order will not be a problem across hosts (use network order)
func (p *packet) encode() (err error) {
	// Reset the buffer in case it has already been written or a previous attempt failed
	p.buffer.Reset()
	var length uint32
	payloadBuffer := bytes.Buffer{}
	// Encode the variable header and payload in temporary buffers so that we know their length,
	// but make sure we only include the payload if there is supposed to be one (and is one)
	vheaderBytes, err := p.properties.Encode()
	if err != nil {
		return err
	}
	length += uint32(len(vheaderBytes))
	if p.payload != nil {
		encoder := gob.NewEncoder(&payloadBuffer)
		err = encoder.Encode(p.payload)
		if err != nil {
			return err
		}
		length += uint32(payloadBuffer.Len())
	}

	// Write the fixed header
	control := p.ptype | p.pflags
	p.buffer.WriteByte(control)
	encodeRemainingLength(p, length)

	// Add the variable header
	p.buffer.Write(vheaderBytes)

	// Add the payload if there is one
	if p.payload != nil {
		p.buffer.Write(payloadBuffer.Bytes())
	}

	return err
}

// decode attempts to populate the fields in the packet by deserializing the encoded slice of
// bytes passed in as a function argument.
func (p *packet) decode(buffer []byte) (err error) {
	return err
}

// initConnect uses the provided properties and payload to initialize the packet as a CONNECT control packet ready to
// be encoded and sent over the network.
func (p *packet) initConnect(properties ConnectProperties, payload ConnectPayload) {
	p.buffer.Reset()
	p.length = 0
	p.ptype = ptypeConnect
	p.pflags = pflagsConnect
	p.properties = properties
	p.payload = payload
}

// initConnectAck uses the provided properties to initialize the packet as a CONNACK control packet ready to be
// encoded and sent over the network
func (p *packet) initConnectAck(properties ConnectAckProperties) {
	p.buffer.Reset()
	p.length = 0
	p.ptype = ptypeConnack
	p.pflags = pflagsConnack
	p.properties = properties
	p.payload = nil
}

// initPublish uses the provided properties and payload to initilaize the packet as a PUBLISH control packet ready
// to be encoded and sent over the network
func (p *packet) initPublish(properties PublishProperties, payload interface{}) {
	p.buffer.Reset()
	p.length = 0
	p.ptype = ptypePublish
	var pflags byte = 0x00
	if properties.DupFlag {
		pflags |= byte(0x08)
	}
	plags |= byte(properties.QoSLevel)
	if properties.Retain {
		pflags |= byte(0x01)
	}
	p.pflags = pflags
	p.properties = properties
	p.payload = payload
}

// initPublishAck uses the provided properties to initialize the packet as a PUBACK control packet ready to be encoded
// and sent over the network
func (p *packet) initPublishAck(properties PublishAckProperties) {
	p.buffer.Reset()
	p.length = 0
	p.ptype = ptypePuback
	p.pflags = pflagsPuback
	p.properties = properties
	p.payload = nil
}

// initPublishRec uses the provided properties to initialize the packet as a PUBREC control packet ready to be encoded
// and sent over the network
func (p *packet) initPublishRec(properties PublishRecProperties) {
	p.buffer.Reset()
	p.length = 0
	p.ptype = ptypePubrec
	p.pflags = pflagsPubrec
	p.properties = properties
	p.payload = nil
}

// initPublishRel uses the provided properties to initialize the packet as a PUBREL control packet ready to be encoded
// and sent over the network
func (p *packet) initPublishRel(properties PublishRelProperties) {
	p.buffer.Reset()
	p.length = 0
	p.ptype = ptypePubrel
	p.pflags = pflagsPubrel
	p.properties = properties
	p.payload = nil
}
