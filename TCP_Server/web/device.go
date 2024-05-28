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

func CreateMockDevices() map[string]*Device {
	devs := make(map[string]*Device, 0)
	devs["225-A-883"] = &Device{
		ID:    "225-A-883",
		Name:  "Vjetroelektrana Kojsino",
		Auth:  "112233",
		State: 1,
	}
	devs["123-C-789"] = &Device{
		ID:    "123-C-789",
		Name:  "Piramida Visoko",
		Auth:  "987654",
		State: 0,
	}
	devs["456-D-012"] = &Device{
		ID:    "456-D-012",
		Name:  "Postaja Zagreb",
		Auth:  "456789",
		State: 0,
	}
	devs["789-E-345"] = &Device{
		ID:    "789-E-345",
		Name:  "Elektrana Zelena Snaga",
		Auth:  "321654",
		State: 1,
	}
	return devs
}
