package wavemq

import (
	"bytes"
	"encoding/gob"
)

// SubscribeChannel ...
type SubscribeChannel struct {
	topic   *Topic
	decoder *gob.Decoder
	buf     bytes.Buffer
	asynch  bool
}

// NewSubscribeChannel ...
func NewSubscribeChannel(t *Topic) *SubscribeChannel {
	ch := SubscribeChannel{topic: t}
	_, ok := t.Message.(Encodeable)
	if !ok {
		ch.decoder = gob.NewDecoder(&ch.buf)
	}
	return &ch
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
	topic   *Topic
	encoder *gob.Encoder
	buf     *bytes.Buffer
	asynch  bool
}

// NewPublishChannel ...
func NewPublishChannel(t *Topic) *PublishChannel {
	ch := PublishChannel{topic: t}
	_, ok := t.Message.(Encodeable)
	if !ok {
		ch.encoder = gob.NewEncoder(&ch.buf)
	}
	return &ch
}

// Send ...
func (pc *PublishChannel) Send(message interface{}) {

}
