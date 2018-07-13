package wavemq

import (
	"bytes"
	"errors"
	"regexp"
	"unicode"
	"unicode/utf8"
)

// ---------------------------------------------------------------------------------------------------------------------
// Constants

// The following constants define the control packet types and their assocaited flags for the MQTT 3.1.1 protocol.
// The control type together with the flag represents the first byte if each packet sent in MQTT. In order to
// create this first byte, WaveMQ will perform a bitwise OR operation on the two bytes representing the type and
// flags respectively.
//
// In accordance with the specification, these values are constant so that WaveMQ can determine if invalid flags
// are attached to each reserved control packet.
//
// REQ: MQTT-2.2.2-1
const (
	// ptypeConnect is the control packet type for client request to connect to a server
	ptypeConnect byte = 0x10
	// ptypeConnack is the control packet type for connection acknowledgment by the server
	ptypeConnack byte = 0x20
	// ptypePublish is the control packet type for a publish message sent from client to server or server to client
	ptypePublish byte = 0x30
	// ptypePuback is the control packet for a PUBACK. Can be sent from client to server or  server to client.
	ptypePuback byte = 0x40
	// ptypePubrec is the publish received control packet type (part one of assured delivery). Sent by the client or
	// server.
	ptypePubrec byte = 0x50
	// ptypePubrel is the publish released control packet type (part two of assured delivery). Sent by the client or
	// server.
	ptypePubrel byte = 0x60
	// ptypePubcomp is the publish complete control packet type (part three of assured delivery). Sent by client or
	// server.
	ptypePubcomp byte = 0x70
	// ptypeSubscribe is the control packet for a client subscrbe request. Sent by the client.
	ptypeSubscribe byte = 0x80
	// ptypeSuback is the control packet type for a client subscription acknowledgment. Sent by the server.
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

// ---------------------------------------------------------------------------------------------------------------------
// Packet Definition

// packet represents a MQTT packet that is either sent or received.
type packet struct {
	ptype      byte
	pflags     byte
	length     uint32
	properties Encodeable
	payload    []byte
	buffer     bytes.Buffer
}

// ---------------------------------------------------------------------------------------------------------------------
// Packet Properties (variable headers), fixed payloads, and encoding implementations

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
	//Decode([]byte) (Encodeable, error)
	// TODO: implement the Decode() function
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
	err = writeIfValidUtf8(&buffer, h.ProtocolName, true)
	if err != nil {
		return nil, err
	}

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
	WillMessage interface{}
	UserName    string
	Password    []byte
}

// Encode writes the payload content for a CONNECT control packet, which has a specific format. This is an
// implementation of the Encodeable interface.
func (p ConnectPayload) Encode() ([]byte, error) {
	buffer := bytes.Buffer{}

	// Encode the identifier after verifying it is valid
	matched, err := regexp.MatchString("[^A-Za-z0-9]+", p.Identifier)
	if matched || err != nil {
		return nil, errors.New("Client identifier must only contain characters A-Z, a-z, or a number")
	} else if l := len(p.Identifier); 1 <= l && l <= 23 {
		return nil, errors.New("Client identifier must be between 1 and 23 bytes")
	}
	err = writeIfValidUtf8(&buffer, p.Identifier, true)
	if err != nil {
		return nil, err
	}

	// Encode the will topic
	if len(p.WillTopic) != 0 {
		err = writeIfValidUtf8(&buffer, p.WillTopic, true)
		if err != nil {
			return nil, err
		}
	}

	// Encode the will message
	if len(p.WillMessage) != 0 {
		err = writeIfValidUtf8(&buffer, p.WillMessage, true)
		if err != nil {
			return nil, err
		}
	}

	// Encode the user name
	if !utf8.ValidString(p.UserName) {
		return nil, errors.New("Invalid UTF-8 encoded user name")
	}
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

// PublishCompProperties defines the fields of the variable header for a PUBCOMP packet.
type PublishCompProperties struct {
	PacketID uint16
}

// Encode writes the variable header of the PUBCOMP message to a byte buffer using the fields and values from the
// PublishCompProperties struct. This is an implementation of the Encodeable interface.
func (h PublishCompProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	buffer.WriteByte(byte(h.PacketID & 0xF0))
	buffer.WriteByte(byte(h.PacketID & 0x0F))
	buf = buffer.Bytes()
	return buf, err
}

// SubscribeProperties defines the fields of the variable header for a SUBSCRIBE control packet.
type SubscribeProperties struct {
	PacketID uint16
}

// Encode writes the variable header of the SUBSCRIBE message to a byte buffer using the fields and values from the
// SubscribeProperties struct. This is an implementation of the Encodeable interface.
func (h SubscribeProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	buffer.WriteByte(byte(h.PacketID & 0xF0))
	buffer.WriteByte(byte(h.PacketID & 0x0F))
	buf = buffer.Bytes()
	return buf, err
}

// SubscribePayload defines the payload of a SUBSCRIBE packet
type SubscribePayload struct {
	Topics map[string]QoSLevel
}

// Encode writes the payload of the SUBSCRIBE message to a byte buffer using the fields and values from the
// SubscribePayload struct. This is an implementation of the Encodeable interface.
func (p SubscribePayload) Encode() (buf []byte, err error) {
	if len(p.Topics) == 0 {
		err = errors.New("SUBSCRIBE payload must have at least one topic/quality of service pair")
		return nil, err
	}
	buffer := bytes.Buffer{}
	for topic, qos := range p.Topics {
		if !utf8.ValidString(topic) {
			return nil, errors.New("Invalid UTF-8 encoded topic")
		}
		buf = []byte(topic)
		buflen := uint16(len(buf))
		buffer.WriteByte(byte(buflen & 0xF0))
		buffer.WriteByte(byte(buflen & 0x0F))
		buffer.Write(buf)
		buffer.WriteByte(byte(qos) >> 1)
	}
	return buffer.Bytes(), err
}

// SubscribeAckProperties defines the fields of the variable header for a SUBACK control packet.
type SubscribeAckProperties struct {
	PacketID uint16
}

// Encode writes the variable header of the SUBACK message to a byte buffer using the fields and values from the
// SubscribeAckProperties struct. This is an implementation of the Encodeable interface.
func (h SubscribeAckProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	buffer.WriteByte(byte(h.PacketID & 0xF0))
	buffer.WriteByte(byte(h.PacketID & 0x0F))
	buf = buffer.Bytes()
	return buf, err
}

// SubscribeAckPayload defines the payload of the SUBACK packet, which comprises of a list of topics and their
// quality of service levels matching the ones sent in the original SUBSCRIBE request.
type SubscribeAckPayload struct {
	Topics map[string]QoSLevel
}

// Encode writes the payload of the SUBACK message to a byte buffer using the fields and values from the
// SubscribeAckPayload struct. This is an implementation of the Encodeable interface.
func (p SubscribeAckPayload) Encode() (buf []byte, err error) {
	if len(p.Topics) == 0 {
		err = errors.New("SUBACK payload must have at least one topic/quality of service pair")
		return nil, err
	}
	buffer := bytes.Buffer{}
	for topic, qos := range p.Topics {
		if !utf8.ValidString(topic) {
			return nil, errors.New("Invalid UTF-8 encoded topic")
		}
		buf = []byte(topic)
		buflen := uint16(len(buf))
		buffer.WriteByte(byte(buflen & 0xF0))
		buffer.WriteByte(byte(buflen & 0x0F))
		buffer.Write(buf)
		buffer.WriteByte(byte(qos) >> 1)
	}
	return buffer.Bytes(), err
}

// UnsubscribeProperties defines the fields of the variable header for a UNSUBSCRIBE control packet.
type UnsubscribeProperties struct {
	PacketID uint16
}

// Encode writes the variable header of the UNSUBSCRIBE message to a byte buffer using the fields and values from the
// SubscribeProperties struct. This is an implementation of the Encodeable interface.
func (h UnsubscribeProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	buffer.WriteByte(byte(h.PacketID & 0xF0))
	buffer.WriteByte(byte(h.PacketID & 0x0F))
	buf = buffer.Bytes()
	return buf, err
}

// UnsubscribePayload defines the payload of a UNSUBSCRIBE packet
type UnsubscribePayload struct {
	Topics map[string]QoSLevel
}

// Encode writes the payload of the UNSUBSCRIBE message to a byte buffer using the fields and values from the
// UnsubscribePayload struct. This is an implementation of the Encodeable interface.
func (p UnsubscribePayload) Encode() (buf []byte, err error) {
	if len(p.Topics) == 0 {
		err = errors.New("SUBSCRIBE payload must have at least one topic/quality of service pair")
		return nil, err
	}
	buffer := bytes.Buffer{}
	for topic, qos := range p.Topics {
		if !utf8.ValidString(topic) {
			return nil, errors.New("Invalid UTF-8 encoded topic")
		}
		buf = []byte(topic)
		buflen := uint16(len(buf))
		buffer.WriteByte(byte(buflen & 0xF0))
		buffer.WriteByte(byte(buflen & 0x0F))
		buffer.Write(buf)
		buffer.WriteByte(byte(qos) >> 1)
	}
	return buffer.Bytes(), err
}

// UnsubscribeAckProperties defines the fields of the variable header for a UNSUBACK control packet.
type UnsubscribeAckProperties struct {
	PacketID uint16
}

// Encode writes the variable header of the UNSUBACK message to a byte buffer using the fields and values from the
// UnsubscribeAckProperties struct. This is an implementation of the Encodeable interface.
func (h UnsubscribeAckProperties) Encode() (buf []byte, err error) {
	buffer := bytes.Buffer{}
	buffer.WriteByte(byte(h.PacketID & 0xF0))
	buffer.WriteByte(byte(h.PacketID & 0x0F))
	buf = buffer.Bytes()
	return buf, err
}

// ---------------------------------------------------------------------------------------------------------------------
// Whole Packet Encoding/Decoding

// writeIfValidUtf8 will take a string and verify that it complies with UTF-8 encoding rules as defined in the Unicode
// spec and RC3629 before writing it to the provided buffer. It ensures that the string will comply with the MQTT \
// protocol, returning an error if the string is not properly encoded or contains the null character (U-000).
// REQ: MQTT-1.5.3-1
func writeIfValidUtf8(buf *bytes.Buffer, s string, writeLength bool) error {
	if writeLength {
		length := uint16(len(s))
		buf.WriteByte(byte(length & 0xF0))
		buf.WriteByte(byte(length & 0x0F))
	}
	for c := range s {
		r := rune(c)
		if !utf8.ValidRune(r) {
			return errors.New("Invalid UTF-8 encoded string")
		} else if r == 0 {
			return errors.New("The encoding of the NULL character (U-000) is not allowed in MQTT")
		} else if r <= 31 || (127 <= r && r <= 159) {
			return errors.New("UTF-8 control characters are not allowed in MQTT")
		}
		buf.WriteRune(r)
	}
	return nil
}

// writeInterface will write the struct passed as the value paramter to the provided buffer. The value interface MAY
// implement the Encodeable interface, in which case it will have an Encode() function defined on it. If this is the
// case, this function will encode the value using that function. If there is no Encode() function defined on the
// provided value, then this function will use the "encoding/gob" package to encode the value.
//
// REQ: MQTT-1.5.3-1
func writeInterface(buf *bytes.Buffer, value interface{}) error {
	return nil
}

// readIfValidUtf8 will take a bytes Buffer and the length it should read from the buffer and verifies that any bytes
// it reads are valid runes (UTF-8 encoded characters) and are not the null character (U-000). If it encounters an
// invalid runes, then it will return the empty string and an error. Otherwise it will return a string containing the
// read values and nil for the error.
//
// REQ: MQTT-1.5.3-1
func readIfValidUtf8(buf *bytes.Buffer, size int) (string, error) {
	runes := make([]rune, 1)
	for size > 0 {
		r, s, err := buf.ReadRune()
		if err != nil {
			return "", err
		} else if r == 0 {
			return "", errors.New("The encoding the the NULL character (U-000) is not allowed in MQTT")
		} else if r == unicode.ReplacementChar {
			return "", errors.New("Invalid UTF-8 encoded rune encountered")
		} else if r <= 31 || (127 <= r && r <= 159) {
			return "", errors.New("UTF-8 control characters are not allowed in MQTT")
		}
		runes = append(runes, r)
		size -= s
	}
	return string(runes), nil
}

// encodeRemainingLength operates on a pointer a packet struct by modifying its internal buffer
// to contain the provided length value in the encoded format specified in the MQTT protocol
// specifications. It will also update the internal offset of the packet so that the rest of
// the packet can be created. This function should only be called when building a packet to send
func encodeRemainingLength(length uint32) []byte {
	buf := make([]byte, 0)
	var encoded byte
	for length > 0 {
		encoded = byte(length % 0x80)
		length /= 0x80
		if length > 0 {
			encoded |= 0x80
		}
		buf = append(buf, encoded)
	}
	return buf
}

// decodeRemainingLength looks at a packet's internal buffer and decodes the value of the
// remaining length s. It also updates the internal read offset by first setting it to 1 (the
// expected location of the start of the remaining length s) and incrementing it so that it
// ends at the start of the variable length header or payload (depending on the packet type)
func decodeRemainingLength(buf []byte) (value uint32, err error) {
	var multiplier uint32 = 1
	var encoded byte
	var index = 0
	first := true
	for (encoded&0x80) != 0 || first {
		if first {
			first = false
		}
		encoded = buf[index]
		index++
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
	// Encode the variable header and payload in temporary buffers so that we know their length,
	// but make sure we only include the payload if there is supposed to be one (and is one)
	vheaderBytes, err := p.properties.Encode()
	if err != nil {
		return err
	}
	length += uint32(len(vheaderBytes))
	if p.payload != nil {
		length += uint32(len(p.payload))
	}

	// Write the fixed header
	control := p.ptype | p.pflags
	p.buffer.WriteByte(control)
	l := encodeRemainingLength(length)
	p.buffer.Write(l)

	// Add the variable header
	p.buffer.Write(vheaderBytes)

	// Add the payload if there is one
	if p.payload != nil {
		p.buffer.Write(p.payload)
	}

	return err
}

// decode attempts to populate the fields in the packet by deserializing the encoded slice of
// bytes passed in as a function argument.
func (p *packet) decode(buffer []byte) (err error) {
	return err
}

// ---------------------------------------------------------------------------------------------------------------------
// Packet Construction/Initialization

// newPacketConnect creates a new CONNECT packet ready to be encoded and sent over the network
func newPacketConnect(properties ConnectProperties, payload ConnectPayload) *packet {
	return &packet{ptype: ptypeConnect, pflags: pflagsConnect, properties: properties, payload: payload.Encode()}
}

// newPacketConnectAck creates a new CONNECT packet ready to be encoded and sent over the network
func newPacketConnectAck(properties ConnectAckProperties) *packet {
	return &packet{ptype: ptypeConnack, pflags: pflagsConnack, properties: properties}
}

// newPacketPublish creates a new PUBLISH packet ready to be encoded and sent over the network
func newPacketPublish(properties PublishProperties, payload []byte) *packet {
	flags := byte(properties.QoSLevel)
	if properties.DupFlag {
		flags |= 0x08
	}
	if properties.Retain {
		flags |= 0x01
	}
	return &packet{ptype: ptypePublish, pflags: flags, properties: properties, payload: payload}
}

// newPacketPublishAck creates a new PUBACK packet ready to be encoded and sent over the network
func newPacketPublishAck(properties PublishAckProperties) *packet {
	return &packet{ptype: ptypePuback, pflags: pflagsPuback, properties: properties, payload: nil}
}

// newPacketPublishRec creates a new PUBREC packet ready to be encoded and sent over the network
func newPacketPublishRec(properties PublishRecProperties) *packet {
	return &packet{ptype: ptypePubrec, pflags: pflagsPubrec, properties: properties}
}

// newPacketPublishRel creates a new PUBREL packet ready to be encoded and sent over the network
func newPacketPublishRel(properties PublishRelProperties) *packet {
	return &packet{ptype: ptypePubrel, pflags: pflagsPubrel, properties: properties}
}

// newPacketPublishComp creates a new PUBCOMP packet ready to be encoded and sent over the network
func newPacketPublishComp(properties PublishCompProperties) *packet {
	return &packet{ptype: ptypePubcomp, pflags: pflagsPubcomp, properties: properties}
}

// newPacketSubscribe creates a new SUBSCRIBE packet ready to be encoded and sent over the network
func newPacketSubscribe(properties SubscribeProperties, payload SubscribePayload) *packet {
	return &packet{ptype: ptypeSubscribe, pflags: pflagsSubscribe, properties: properties, payload: payload.Encode()}
}

// newPacketSubscribeAck creates a new SUBACK packet ready to be encoded and sent over the network
func newPacketSubscribeAck(properties SubscribeAckProperties, payload SubscribeAckPayload) *packet {
	return &packet{ptype: ptypeSuback, pflags: pflagsSuback, properties: properties, payload: payload.Encode()}
}

// newPacketSubscribe creates a new UNSUBCRIBE packet ready to be encoded and sent over the network
func newPacketUnsubscribe(properties UnsubscribePayload, payload UnsubscribePayload) *packet {
	return &packet{ptype: ptypeUnsubscribe, pflags: pflagsUnsubscribe, properties: properties, payload: payload.Encode()}
}

// newPacketUnsubscribeAck creates a new UNSUBACK packet ready to be encoded and sent over the network
func newPacketUnsubscribeAck(properties UnsubscribeAckPayload) *packet {
	return &packet{ptype: ptypeUnsuback, pflags: pflagsUnsuback, properties: properties}
}

// newPacketPingReq creates a new PINGREQ packet ready to be encoded and sent over the network
func newPacketPingReq() *packet {
	return &packet{ptype: ptypePingreq, pflags: pflagsPingreq}
}

// newPacketPingResp creates a new PINGRESP packet ready to be encoded and sent over the network
func newPacketPingResp() *packet {
	return &packet{ptype: ptypePingresp, pflags: pflagsPingresp}
}

// newPacketDisconnect creates a new DISCONNECT packet ready to be encoded and sent over the network
func newPacketDisconnect() {
	return &packet{ptype: ptypeDisconnect, pflags: pflagsDisconnect}
}
