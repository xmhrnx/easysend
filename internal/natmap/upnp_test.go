package natmap

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const desc = `<?xml version="1.0"?><root><device><deviceList><device><deviceList><device>
<serviceList><service><serviceType>urn:schemas-upnp-org:service:WANIPConnection:1</serviceType>
<controlURL>/ctl/IPConn</controlURL></service></serviceList></device></deviceList></device></deviceList></device></root>`

func TestParseAndSOAP(t *testing.T) {
	var actions []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		actions = append(actions, r.Header.Get("SOAPAction"))
		if strings.Contains(string(b), "GetExternalIPAddress") {
			io.WriteString(w, `<s:Envelope xmlns:s="x"><s:Body><u:GetExternalIPAddressResponse xmlns:u="y"><NewExternalIPAddress>203.0.113.7</NewExternalIPAddress></u:GetExternalIPAddressResponse></s:Body></s:Envelope>`)
		}
	}))
	defer srv.Close()
	gw, err := parseDescription(srv.URL+"/desc.xml", []byte(desc))
	if err != nil || gw.controlURL != srv.URL+"/ctl/IPConn" {
		t.Fatalf("%v %+v", err, gw)
	}
	gw.localIP = "192.168.1.5"
	m := &Mapping{InternalPort: 1, ExternalPort: 1, gw: gw}
	if err := m.add(context.Background(), 0, "x"); err != nil {
		t.Fatal(err)
	}
	ip, err := gw.externalIP(context.Background())
	if err != nil || ip != "203.0.113.7" {
		t.Fatal(ip, err)
	}
	m.ExternalIP = ip
	if !m.IsPublic() || !strings.HasSuffix(actions[0], "#AddPortMapping\"") {
		t.Fatal(actions)
	}
}
