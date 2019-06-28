package sonar

type Service struct {
	Type string `json:"type"`
	Protocol string `json:"protocol"`
	Address string `json:"address"`
	Port int `json:"port"`
}

type Message struct {
	Type int           `json:"type"`
	ID int             `json:"id"`
	Domain string      `json:"domain,omitempty"`
	Requestor string   `json:"requestor,omitempty"`
	Services []Service `json:"services,omitempty"`
}

const (
	SonarPing = iota
	SonarEcho
)

const (
	DefaultDomain = "nano"
	DefaultMulticastAddress = "224.0.0.226"
	DefaultMulticastPort = 5599
	DefaultBufferSize = 8 << 10
)