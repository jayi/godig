package godig

import (
	"github.com/miekg/dns"
	"time"
	"net"
	"fmt"
	"strings"
)

const defaultServer = "119.29.29.29"

type DigClient struct {
	Server string
	Port int
	Domain string
	Client string
	Retries int
	Timeout time.Duration
}

func New() *DigClient {
	return &DigClient{
		Server: defaultServer,
		Port: 53,
		Timeout: time.Second * 5,
		Retries: 3,
	}
}

func (client *DigClient) SetServer(server string) *DigClient {
	client.Server = server
	return client
}

func (client *DigClient) SetPort(port int) *DigClient {
	client.Port = port
	return client
}

func (client *DigClient) SetRetries(retires int) *DigClient {
	client.Retries = retires
	return client
}

func (client *DigClient) SetClient(subnet string) *DigClient {
	client.Client = subnet
	return client
}

func (client *DigClient) SetTimeout(timeout time.Duration) *DigClient {
	client.Timeout = timeout
	return client
}

func (client *DigClient) QueryDomain(domain string) (r *dns.Msg, rtt time.Duration, err error) {
	msg := new(dns.Msg)
	msg.SetQuestion(domain + ".", dns.TypeA)

	if client.Client != "" {
		opt := new(dns.OPT)
		opt.Hdr.Name = "."
		opt.Hdr.Rrtype = dns.TypeOPT
		e := new(dns.EDNS0_SUBNET)
		e.Code = dns.EDNS0SUBNET
		e.Family = 1 // 1 for IPv4 source address, 2 for IPv6
		e.SourceNetmask = 32 // 32 for IPV4, 128 for IPv6
		e.SourceScope = 0
		e.Address = net.ParseIP(client.Client).To4()
		opt.Option = append(opt.Option, e)
		msg.Extra = append(msg.Extra, opt)
	}

	c := new(dns.Client)
	c.Timeout = client.Timeout

	addr := fmt.Sprintf("%s:%d", client.Server, (client.Port))
	for i := 0; i < client.Retries + 1; i++ {
		r, rtt, err = c.Exchange(msg, addr)
		if err == nil {
			return
		}

		shouldRetry := false
		if strings.Contains(err.Error(), dns.ErrTruncated.Error()) {
			c.Net = "tcp"
			shouldRetry = true
		}
		netErr, ok := err.(net.Error)
		if ok && netErr.Temporary() {
			shouldRetry = true
		}

		if !shouldRetry {
			return
		}
		time.Sleep(time.Second)
	}
	return
}

