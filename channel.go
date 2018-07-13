package wavemq

import (
	"bytes"
	"encoding/gob"
	"errors"
)

// AsynchAction describes the signature of the function that will be used to handle asynchronous subscriptions. It
// keeps the developer from registering any kind of function they want, restricting them to a signature that is
// compatable with WaveMQ.
type AsynchAction func(interface{})

// Subscriber defines the member properties of a subscriber in WaveMQ. The subscriber is responsible for retrieving
// messages from the broker for the topic it has subscribed to and then handing it off to the application.
type Subscriber struct {
	topic   *Topic
	decoder *gob.Decoder
	buf     bytes.Buffer
	asynch  bool
	action  AsynchAction
}

// NewSubscriber creates a traditional, synchronous subscriber on the provided topic and returns a pointer to it.
func NewSubscriber(t *Topic) *Subscriber {
	sub := Subscriber{topic: t}
	_, ok := t.Message.(Encodeable)
	if !ok {
		sub.decoder = gob.NewDecoder(&sub.buf)
	}
	return &sub
}

// NewAsyncSubscriber creates a new asynchronous subscriber. This type of subscriber will periodically attempt to read
// from the broker (via a golang channel) and, whenever data is detected, will immediately decode the message and
// invoke the action registered with the channel.
func NewAsyncSubscriber(t *Topic, action AsynchAction) {
	// TODO: ACTUALLY implement reading from the channel
	sub := Subscriber{topic: t, asynch: true, action: action}
	if _, ok := t.Message.(Encodeable); !ok {
		sub.decoder = gob.NewDecoder(&sub.buf)
	}
	return &sub
}

// ReceiveIn ...
func (sc *Subscriber) ReceiveIn(target interface{}) {

}

// Publisher ...
type Publisher struct {
	Properties PublishProperties
	topic      *Topic
	encoder    *gob.Encoder
	buf        *bytes.Buffer
	asynch     bool
}

// NewPublisher ...
func NewPublisher(t *Topic) *Publisher {
	ch := Publisher{topic: t}
	_, ok := t.Message.(Encodeable)
	if !ok {
		ch.encoder = gob.NewEncoder(&ch.buf)
	}
	return &ch
}

// Send ...
func (pc *Publisher) Send(message interface{}) error {
	var payload []byte
	if pc.encoder != nil {
		payload = pc.encoder.Encode(&message)
	} else if _, ok := message.(Encodeable); ok {
		payload = message.Encode()
	} else {
		return errors.New("Unable to encode message because it does not implement 'Encodeable' and no encode is found")
	}
	p := packet{}
	p.initPublish(pc.Properties, payload)

	// TODO: send the packet
	return nil
}
