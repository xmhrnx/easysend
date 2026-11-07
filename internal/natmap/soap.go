package natmap

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type kv struct{ k, v string }

func (g *gateway) soap(ctx context.Context, action string, args []kv) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0"?><s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><s:Body>`)
	fmt.Fprintf(&b, `<u:%s xmlns:u="%s">`, action, g.serviceType)
	for _, a := range args {
		b.WriteString("<" + a.k + ">")
		_ = xml.EscapeText(&b, []byte(a.v))
		b.WriteString("</" + a.k + ">")
	}
	fmt.Fprintf(&b, `</u:%s></s:Body></s:Envelope>`, action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.controlURL, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", `"`+g.serviceType+"#"+action+`"`)
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("router refused %s (%s)", action, resp.Status)
	}
	return body, nil
}

func (g *gateway) externalIP(ctx context.Context) (string, error) {
	body, err := g.soap(ctx, "GetExternalIPAddress", nil)
	if err != nil {
		return "", err
	}
	var env struct {
		IP string `xml:"Body>GetExternalIPAddressResponse>NewExternalIPAddress"`
	}
	if err := xml.Unmarshal(body, &env); err != nil {
		return "", err
	}
	if env.IP == "" {
		return "", errors.New("router didn't report its address")
	}
	return env.IP, nil
}
