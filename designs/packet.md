The packet.go file contains primarily non-exported types and functions

```golang
package wavemq

import(
    "bytes"
)

type packet struct {
    ptype byte
    pflags byte
    length uint32
    properties Encodeable
    payload []byte // encode/decoded by channel
    buffer bytes.Buffer
}

// Creating a packet
p := packet{}
p.initConnect()
// OR
p := newPacketConnect()

// Uses reflection to determine how to encode the packet. Populates the buffer
p.encode()

// Uses the first byte of the buffer to determine how to decode the rest of the buffer
buffer := bytes.Buffer{}
p.decode(&buffer)
```