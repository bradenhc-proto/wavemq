The Topic struct has several useful functions and fields that make publishing and subscribing simpler. This includes
helper functions that encode/decode the messages on that topic.


```golang
// Simplest implementation that will use the GOB encoder to encode/decode the message
type MyGobMessage struct {
    to string
    from string
    body string
}

t := Topic{name: "my-topic", message: MyGobMessage{}}

//
// The structure can also implement the Encodable interface by defining an Encode() and Decode() function on the
// structure. When the Encodeable interface is implemented, those functiona will be used instead of the GOB encoder
// to encode/decode the messages
type MyCustomMessage struct {
    to string
    from string
    body string
}

func (m *MyCustomMessage) Encode() ([]byte, error) {
    // implementation here
}

func (m *MyCustomMessage) Decode([]byte) (error) {
    // implementation here
}

// This topic will now encode/decode using the custom implementations of the Encodeable interface
t := Topic{name: "my-topic", message: MyCustomMessage{}}
```

Inside of the `PublishChannel` and `SubscribeChannel`, WaveMQ will use the appropriate encoder defined inside the
`Topic` when sending/receiving messages synchronously or asynchronously.