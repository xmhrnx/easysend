package natmap

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type xmlService struct {
	ServiceType string `xml:"serviceType"`
	ControlURL  string `xml:"controlURL"`
}

type xmlDevice struct {
	Services []xmlService `xml:"serviceList>service"`
	Devices  []xmlDevice  `xml:"deviceList>device"`
}

type xmlRoot struct {
	URLBase string    `xml:"URLBase"`
	Device  xmlDevice `xml:"device"`
}

func describe(ctx context.Context, loc string) (*gateway, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loc, nil)
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 4 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	gw, err := parseDescription(loc, body)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(loc)
	gw.localIP = localIPFor(u.Host)
	return gw, nil
}

func parseDescription(loc string, body []byte) (*gateway, error) {
	var root xmlRoot
	if err := xml.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	svc := findService(root.Device)
	if svc == nil {
		return nil, errors.New("router has no WAN connection service")
	}
	base := root.URLBase
	if base == "" {
		base = loc
	}
	bu, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	cu, err := bu.Parse(svc.ControlURL)
	if err != nil {
		return nil, err
	}
	return &gateway{controlURL: cu.String(), serviceType: svc.ServiceType}, nil
}

func findService(d xmlDevice) *xmlService {
	for i, s := range d.Services {
		if strings.Contains(s.ServiceType, "WANIPConnection") || strings.Contains(s.ServiceType, "WANPPPConnection") {
			return &d.Services[i]
		}
	}
	for _, c := range d.Devices {
		if s := findService(c); s != nil {
			return s
		}
	}
	return nil
}

func localIPFor(host string) string {
	if !strings.Contains(host, ":") {
		host += ":80"
	}
	c, err := net.Dial("udp4", host)
	if err != nil {
		return ""
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).IP.String()
}
