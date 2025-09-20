package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"sync"

	wsdiscovery "github.com/IOTechSystems/onvif/ws-discovery"
	"golang.org/x/sync/errgroup"
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

func DiscoverNewCameras(ctx context.Context, logger *slog.Logger) discovery.WsDiscoveryDto {
	ifs, _ := net.Interfaces()

	var (
		mtx        sync.Mutex
		discovered = discovery.WsDiscoveryDto{Matches: []discovery.WsDiscoveryMatch{}}
	)
	eg, ctx := errgroup.WithContext(ctx)

	for _, iface := range ifs {
		name := iface.Name

		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			responses, err := wsdiscovery.SendProbe(
				name,
				nil,
				nil,
				nil,
			)

			if err != nil {
				logger.Debug("wsdiscovery probe failed", "iface", name, "err", err)
				return nil
			}

			for _, resp := range responses {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				var out wsDiscoveryResp
				if err := xml.Unmarshal([]byte(resp), &out); err != nil {
					logger.Error("xml unmarshal failed", "err", err)
					continue
				}

				for _, m := range out.Matches {
					sub := uuidRE.FindStringSubmatch(m.Match.UUID)
					if len(sub) < 2 {
						logger.Warn(fmt.Sprintf("Could not find a submatch for %s", m.Match.UUID))
						continue
					}
					match := discovery.WsDiscoveryMatch{
						UUID:  sub[1],
						Xaddr: hostPortRE.FindString(m.Match.Xaddr),
					}

					mtx.Lock()
					discovered.Matches = append(discovered.Matches, match)
					mtx.Unlock()
				}
			}

			return nil

		})
	}

	if err := eg.Wait(); err != nil {
		logger.Info("discovery canceled/timeout", "err", err)
	}

	return discovered
}
