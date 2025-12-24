package eventbus

import (
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

type ConnectionStatus string

const (
	Pending      ConnectionStatus = "pending"
	Active       ConnectionStatus = "active"
	Disconnected ConnectionStatus = "disconnected"
)

type NatsConfig struct {
	natsUrl             string
	appName             string
	requiresCredentials bool
	username            string
	password            string
}

type NatsConnInstance struct {
	conn   *nats.Conn
	status ConnectionStatus
}

type NatsConnection interface {
	Status() ConnectionStatus
	Disconnect()
	GetConnection() (*nats.Conn, error)
}

func NewNatsConnection(natsConf NatsConfig) (*NatsConnInstance, error) {

	connInstance := &NatsConnInstance{
		conn:   nil,
		status: Pending,
	}
	url := natsConf.natsUrl
	if natsConf.requiresCredentials {
		// later check if password not passed errro
		if strings.TrimSpace(natsConf.username) == "" {
			return nil, fmt.Errorf("The username is blank (empty or only whitespace)")
		}
		if strings.TrimSpace(natsConf.password) == "" {
			return nil, fmt.Errorf("The password is blank (empty or only whitespace)")
		}
		url = fmt.Sprintf("%s:%s@%s", natsConf.username, natsConf.password, url)
	}

	nc, err := nats.Connect(url,
		nats.Name(natsConf.appName),
		nats.Timeout(30*time.Second),
		nats.MaxReconnects(5),
		nats.ReconnectWait(time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			fmt.Printf("Connection lost: %v\n", err)
			connInstance.status = Disconnected
		}),
	)

	if err != nil {
		return nil, err
	}
	connInstance.conn = nc
	connInstance.status = Active

	return connInstance, nil

}

func (nt *NatsConnInstance) Status() ConnectionStatus {

	return nt.status
}

func (nt *NatsConnInstance) Disconnect() {
	if nt.conn != nil {
		if nt.status == Active {
			nt.conn.Close()
		}

	}

}

func (nt *NatsConnInstance) GetConnection() (*nats.Conn, error) {
	if nt.conn != nil {
		return nt.conn, nil
	}

	return nil, fmt.Errorf("The connection is invalid  status %s", nt.status)
}
