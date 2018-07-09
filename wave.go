package wavemq

// Session ...
type Session struct {
	Name                 string
	ServerAddress        string
	identifier           string
	ConnectionProperties ConnectProperties
	state                interface{}
}
