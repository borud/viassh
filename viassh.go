// Package viassh provides a convenient method for tunneling connections through one
// or more SSH tunnels.
package viassh

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Config for dialer.
type Config struct {
	// Hosts are on the format username@host:port and all components of the host entry are mandatory.
	Hosts  []string
	Logger *log.Logger
}

// Dialer that will tunnel connections through one or more layers of SSH connections.
type Dialer struct {
	clients []*ssh.Client
	config  Config
}

type sshDialerFunc func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)

// errors
var (
	ErrParsingHost   = errors.New("error parsing host entry")
	ErrDialing       = errors.New("error dialing host")
	ErrNoHostEntries = errors.New("no via entries specified")
	ErrSSHAgent      = errors.New("error opening ssh-agent")
)

// Create dialer.
func Create(c Config) (*Dialer, error) {
	if len(c.Hosts) == 0 {
		return nil, ErrNoHostEntries
	}

	if c.Logger == nil {
		c.Logger = log.New(io.Discard, "", 0)
	}

	authSock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSSHAgent, err)
	}

	agentClient := agent.NewClient(authSock)
	clients := make([]*ssh.Client, len(c.Hosts))

	var dialer = ssh.Dial
	for i := 0; i < len(c.Hosts); i++ {
		host := c.Hosts[i]

		c.Logger.Printf("connecting to [%s]", host)

		url, err := url.Parse("ssh://" + host)
		if err != nil {
			return nil, fmt.Errorf("%w %s (%d): %v", ErrParsingHost, host, i, err)
		}

		clientConfig := &ssh.ClientConfig{
			User: url.User.Username(),
			Auth: []ssh.AuthMethod{
				ssh.PublicKeysCallback(agentClient.Signers),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		client, err := dialer("tcp", url.Host, clientConfig)
		if err != nil {
			return nil, fmt.Errorf("%w %s (%d): %v", ErrDialing, host, i, err)
		}
		clients[i] = client
		dialer = nextDialer(client)
	}

	return &Dialer{
		clients: clients,
		config:  c,
	}, err
}

// Dial an address from the remote end and return the connection and possibly an error.
func (d *Dialer) Dial(prot, addr string) (net.Conn, error) {
	d.config.Logger.Printf("dialing [%s]", addr)
	client := d.clients[len(d.clients)-1]
	return client.Dial(prot, addr)
}

func (d *Dialer) Close() error {
	errs := []error{}

	for i := len(d.clients) - 1; i >= 0; i-- {
		d.config.Logger.Printf("closing ssh connection to [%s]", d.config.Hosts[i])
		err := d.clients[i].Close()
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func nextDialer(client *ssh.Client) sshDialerFunc {
	return func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		conn, err := client.Dial(network, addr)
		if err != nil {
			return nil, err
		}

		ncc, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
		if err != nil {
			return nil, err
		}

		return ssh.NewClient(ncc, chans, reqs), nil
	}
}
