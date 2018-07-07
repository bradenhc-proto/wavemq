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
    client := wavemq.Client{}
    client.Connect("192.168.1.124", wavemq.ConnectionProperties{})

    // OR ----------------------------------------------------------------------------------------
    // For an existing session or one saved previously
    session := client.GetSession("key")
    client.Reconnect(session)
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
    subscriptionChannel := client.SubscribeTo(topic, func(message Message){
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