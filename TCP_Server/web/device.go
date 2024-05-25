package web

type Device struct {
	ID    string
	Name  string
	State int
}

func GetDevlist() map[string]*Device {
	return tcpServer.devices
}

func DevName(id string) string {
	return tcpServer.devices[id].Name
}
