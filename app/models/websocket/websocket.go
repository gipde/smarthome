package websocket

type DeviceCommand struct {
	Device     string
	Connected  bool
	Command    string
	State      string
	DeviceType int
}

type StateTopic struct {
	Input    chan string
	Consumer [](chan string)
}
