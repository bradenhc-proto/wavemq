### Example Usage for Developer

```golang
import (
    "github.com/ambientms/wavemq"
)

type Message struct {
    to string
    from string
    content string
}

func main() {
    // Initializing the client and connecting
    // WaveMQ can be configured to associate names with IP addresses in a config file (or using /etc/hosts),
    // so you don't need to pass an IP address to make it work. The string 'localhost' also works. Can also
    // use names of machines on the same network. Default port is already set.
    client := wavemq.Client{}
    client.Connect("192.168.1.124", wavemq.ConnectProperties{})

    // OR ----------------------------------------------------------------------------------------
    // For an existing session or one saved previously
    client.Reconnect("id")
    // -------------------------------------------------------------------------------------------

    topic := wavemq.Topic{name: "my-sub-topic", message: Message{}}
    subscriptionChannel, err := client.SubscribeTo(topic)
    message := Message{}
    subscriptionChannel.ReceiveIn(&message)

    // OR ----------------------------------------------------------------------------------------
    // This way the client has access to the message IMMEDIATELY (relative to other subscribers)
    // after it has been published to the topic. The above method is a queue-based implementation.
    // The below method is an event-based implementation (remove a lot of heavy lifting on the part
    // of the developer). Perhaps leave it up to them to decide??
    // Registering a callback in this way is thread safe up to the point of the subscribing function
    // NOTE: YOU CANNOT SUBSCRIBE TO A TOPIC BOTH SYNCHRONOUSLY AND ASYNCHRONOUSLY. You must choose
    // one or the other. Attempting to do both will result in an error (panic??)
    subscriptionChannel := client.SubscribeToAsynch(topic, func(message Message){
        // Do something with the message here
        // Be sure the argument is of the same type of the topic message, otherwise there
        // will be problems
        // Quality of service will gaurantee that when this function is invoked, there is a
        // object of the appropriate type available that gets passed to the parameter
    });
    subscriptionChannel.Mute() // Pauses processing events
    subscriptionChannel.Unmute() // Resumes processing events
    subscriptionChannel.Close() // Closes the subscription while keeping the client connected
    // -------------------------------------------------------------------------------------------

    topic := wavemq.Topic{name: "my-pub-topic", message: Message{}}
    publishChannel, err := client.PublishOn(topic)
    message := Message{
        to: "The World",
        from: "John Smith",
        content: "Hello World!"
    }
    publishChannel.Send(message)

    client.Close()
}
```