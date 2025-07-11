package agent

import (
	"fmt"
	"net"
	"strconv"

	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/serf/serf"

	sxcfg "github.com/sine-io/sinx/internal/config"
)

var (
	// projectURL is the project URL.
	projectURL = "https://www.sineio.top/"
)

// ServerParts is used to return the parts of a server role
type ServerParts struct {
	Name         string
	ID           string
	Region       string
	Datacenter   string
	Port         int
	Bootstrap    bool
	Expect       int
	RaftVersion  int
	BuildVersion *version.Version
	Addr         net.Addr
	RPCAddr      net.Addr
	Status       serf.MemberStatus
}

// String returns a representation of this instance
func (s *ServerParts) String() string {
	return fmt.Sprintf("%s (Addr: %s) (DC: %s)",
		s.Name, s.Addr, s.Datacenter)
}

// Copy returns a copy of this struct
func (s *ServerParts) Copy() *ServerParts {
	ns := new(ServerParts)
	*ns = *s
	return ns
}

// UserAgent returns the consistent user-agent string
func UserAgent() string {
	return fmt.Sprintf("Sinx/%s (+%s;)", sxcfg.Version, projectURL)
}

// IsServer Returns if a member is a Sinx server. Returns a boolean,
// and a struct with the various important components
func isServer(m serf.Member) (bool, *ServerParts) {
	if m.Tags["role"] != "sinx" {
		return false, nil
	}

	if m.Tags["server"] != "true" {
		return false, nil
	}

	id := m.Name
	region := m.Tags["region"]
	datacenter := m.Tags["dc"]
	_, bootstrap := m.Tags["bootstrap"]

	expect := 0
	expectStr, ok := m.Tags["expect"]
	var err error
	if ok {
		expect, err = strconv.Atoi(expectStr)
		if err != nil {
			return false, nil
		}
	}
	// TODO: ?
	if expect == 1 {
		bootstrap = true
	}

	// If the server is missing the rpc_addr tag, default to the serf advertise addr
	rpcIP := net.ParseIP(m.Tags["rpc_addr"])
	if rpcIP == nil {
		rpcIP = m.Addr
	}

	portStr := m.Tags["port"]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false, nil
	}

	buildVersion, err := version.NewVersion(m.Tags["version"])
	if err != nil {
		buildVersion = &version.Version{}
	}

	addr := &net.TCPAddr{IP: m.Addr, Port: port}
	rpcAddr := &net.TCPAddr{IP: rpcIP, Port: port}
	parts := &ServerParts{
		Name:         m.Name,
		ID:           id,
		Region:       region,
		Datacenter:   datacenter,
		Port:         port,
		Bootstrap:    bootstrap,
		Expect:       expect,
		Addr:         addr,
		RPCAddr:      rpcAddr,
		BuildVersion: buildVersion,
		Status:       m.Status,
	}
	return true, parts
}
