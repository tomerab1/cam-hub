package onvif

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"sync"

	wsdiscovery "github.com/IOTechSystems/onvif/ws-discovery"
	"tomerab.com/cam-hub/internal/onvif/discovery"
)

var hostPortRE = regexp.MustCompile("[0-9]+.+[0-9]+:[0-9]+")
var uuidRE = regexp.MustCompile(`urn:uuid:([0-9a-fA-F-]{36})`)

type wsDiscoveryResp struct {
	Matches []struct {
		Match struct {
			UUID  string `xml:"EndpointReference>Address"`
			Xaddr string `xml:"XAddrs"`
		} `xml:"ProbeMatch"`
	} `xml:"Body>ProbeMatches"`
}

func DiscoverNewCameras(logger *slog.Logger) discovery.WsDiscoveryDto {
	ifs, _ := net.Interfaces()

	var matchesLk sync.Mutex
	var wg sync.WaitGroup
	discovered := discovery.WsDiscoveryDto{Matches: []discovery.WsDiscoveryMatch{}}

	for _, i := range ifs {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			responses, err := wsdiscovery.SendProbe(
				name,
				nil,
				nil,
				nil,
			)

			if err != nil {
				logger.Debug(err.Error(), "iface", name)
				return
			}

			for _, resp := range responses {
				var out wsDiscoveryResp
				if err = xml.Unmarshal([]byte(resp), &out); err != nil {
					logger.Error(err.Error())
					continue
				}

				for _, m := range out.Matches {
					matchesLk.Lock()
					defer matchesLk.Unlock()

					matches := uuidRE.FindStringSubmatch(m.Match.UUID)
					if len(matches) < 2 {
						logger.Warn(fmt.Sprintf("Could not find a submatch for %s", m.Match.UUID))
						continue
					}

					discovered.Matches = append(discovered.Matches, discovery.WsDiscoveryMatch{
						UUID:  uuidRE.FindStringSubmatch(m.Match.UUID)[1],
						Xaddr: hostPortRE.FindString(m.Match.Xaddr),
					})
				}
			}
		}(i.Name)
	}
	wg.Wait()
	return discovered
}
