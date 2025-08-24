package onvif

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type soapEnvelope[T any] struct {
	Body struct {
		Resp  *T         `xml:",any"`
		Fault *SOAPFault `xml:"Fault"`
	} `xml:"Body"`
}

type SOAPFault struct {
	Code struct {
		Value   string `xml:"Value"`
		Subcode struct {
			Value string `xml:"Value"`
		} `xml:"Subcode"`
	} `xml:"Code"`
	Reason struct {
		Text string `xml:"Text"`
	} `xml:"Reason"`
}

func parseResp[T any](resp *http.Response, out *T) error {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	defer resp.Body.Close()

	var env soapEnvelope[T]
	if err := xml.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("failed to unmarshal SOAP envelope: %w", err)
	}

	if env.Body.Fault != nil {
		return fmt.Errorf("SOAP fault: %s", env.Body.Fault.Reason.Text)
	}

	if env.Body.Resp == nil {
		return fmt.Errorf("no response data found in SOAP body")
	}

	*out = *env.Body.Resp
	return nil
}
