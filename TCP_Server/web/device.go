package web

type Device struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Auth  string `json:"auth"`
	State int
}

func GetDevlist() map[string]*Device {
	return tcpServer.devices
}

func DevName(id string) string {
	return tcpServer.devices[id].Name
}
