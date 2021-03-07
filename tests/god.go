package main



import (
	"crypto/sha256"
	"fmt"
	"bytes"
	"io"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"io/ioutil"
	// "golang.org/x/net/http2"
	tls "github.com/refraction-networking/utls"
)



type Browser struct {
	JA3       string
	UserAgent string
}

// ChromeAuto mocks Chrome 78
var ChromeAuto = Browser{
	JA3:       "769,47–53–5–10–49161–49162–49171–49172–50–56–19–4,0–10–11,23–24–25,0",
	UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.97 Safari/537.36",
}

// SafariAuto mocks Safari 604.1
var SafariAuto = Browser{
	JA3:       "771,4865-4866-4867-49196-49195-49188-49187-49162-49161-52393-49200-49199-49192-49191-49172-49171-52392-157-156-61-60-53-47-49160-49170-10,65281-0-23-13-5-18-16-11-51-45-43-10-21,29-23-24-25,0",
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 13_1_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.1 Mobile/15E148 Safari/604.1",
}

// greasePlaceholder is a random value (well, kindof '0x?a?a) specified in a
// random RFC.
const greasePlaceholder = 0x0a0a

// ErrExtensionNotExist is returned when an extension is not supported by the library
type ErrExtensionNotExist string

// Error is the error value which contains the extension that does not exist
func (e ErrExtensionNotExist) Error() string {
	return fmt.Sprintf("Extension does not exist: %s\n", e)
}

// extMap maps extension values to the TLSExtension object associated with the
// number. Some values are not put in here because they must be applied in a
// special way. For example, "10" is the SupportedCurves extension which is also
// used to calculate the JA3 signature. These JA3-dependent values are applied
// after the instantiation of the map.
var extMap = map[string]tls.TLSExtension{
	"0": &tls.SNIExtension{},
	"5": &tls.StatusRequestExtension{},
	// These are applied later
	// "10": &tls.SupportedCurvesExtension{...}
	// "11": &tls.SupportedPointsExtension{...}
	"13": &tls.SignatureAlgorithmsExtension{
		SupportedSignatureAlgorithms: []tls.SignatureScheme{
			tls.ECDSAWithP256AndSHA256,
			tls.PSSWithSHA256,
			tls.PKCS1WithSHA256,
			tls.ECDSAWithP384AndSHA384,
			tls.PSSWithSHA384,
			tls.PKCS1WithSHA384,
			tls.PSSWithSHA512,
			tls.PKCS1WithSHA512,
			tls.PKCS1WithSHA1,
		},
	},
	"16": &tls.ALPNExtension{
		AlpnProtocols: []string{"h2", "http/1.1"},
	},
	"18": &tls.SCTExtension{},
	"21": &tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
	"23": &tls.UtlsExtendedMasterSecretExtension{},
	"27": &tls.FakeCertCompressionAlgsExtension{},
	"28": &tls.FakeRecordSizeLimitExtension{},
	"35": &tls.SessionTicketExtension{},
	"43": &tls.SupportedVersionsExtension{[]uint16{
		tls.GREASE_PLACEHOLDER,
		tls.VersionTLS13,
		tls.VersionTLS12,
		tls.VersionTLS11,
		tls.VersionTLS10}},
	"44": &tls.CookieExtension{},
	"45": &tls.PSKKeyExchangeModesExtension{[]uint8{
		tls.PskModeDHE,
	}},
	"51":    &tls.KeyShareExtension{[]tls.KeyShare{}},
	"13172": &tls.NPNExtension{},
	"65281": &tls.RenegotiationInfoExtension{
		Renegotiation: tls.RenegotiateOnceAsClient,
	},
}

// NewTransport creates an http.Transport which mocks the given JA3 signature when HTTPS is used
func NewTransport(ja3 string) (*http.Transport, error) {
	return NewTransportWithConfig(ja3, &tls.Config{})
}

// NewTransportWithConfig creates an http.Transport object given a utls.Config
func NewTransportWithConfig(ja3 string, config *tls.Config) (*http.Transport, error) {
	spec, err := stringToSpec(ja3)
	if err != nil {
		return nil, err
	}

	dialtls := func(network, addr string) (net.Conn, error) {
		dialConn, err := net.Dial(network, addr)
		if err != nil {
			return nil, err
		}

		config.ServerName = strings.Split(addr, ":")[0]

		uTlsConn := tls.UClient(dialConn, config, tls.HelloCustom)
		if err := uTlsConn.ApplyPreset(spec); err != nil {
			return nil, err
		}
		if err := uTlsConn.Handshake(); err != nil {
			return nil, err
		}
		return uTlsConn, nil
	}

	return &http.Transport{DialTLS: dialtls}, nil
}

// stringToSpec creates a ClientHelloSpec based on a JA3 string
func stringToSpec(ja3 string) (*tls.ClientHelloSpec, error) {
	tokens := strings.Split(ja3, ",")

	version := tokens[0]
	ciphers := strings.Split(tokens[1], "-")
	extensions := strings.Split(tokens[2], "-")
	curves := strings.Split(tokens[3], "-")
	if len(curves) == 1 && curves[0] == "" {
		curves = []string{}
	}
	pointFormats := strings.Split(tokens[4], "-")
	if len(pointFormats) == 1 && pointFormats[0] == "" {
		pointFormats = []string{}
	}

	// parse curves
	var targetCurves []tls.CurveID
	for _, c := range curves {
		cid, err := strconv.ParseUint(c, 10, 16)
		if err != nil {
			return nil, err
		}
		targetCurves = append(targetCurves, tls.CurveID(cid))
	}
	extMap["10"] = &tls.SupportedCurvesExtension{targetCurves}

	// parse point formats
	var targetPointFormats []byte
	for _, p := range pointFormats {
		pid, err := strconv.ParseUint(p, 10, 8)
		if err != nil {
			return nil, err
		}
		targetPointFormats = append(targetPointFormats, byte(pid))
	}
	extMap["11"] = &tls.SupportedPointsExtension{SupportedPoints: targetPointFormats}

	// build extenions list
	var exts []tls.TLSExtension
	for _, e := range extensions {
		te, ok := extMap[e]
		if !ok {
			return nil, ErrExtensionNotExist(e)
		}
		exts = append(exts, te)
	}
	// build SSLVersion
	vid64, err := strconv.ParseUint(version, 10, 16)
	if err != nil {
		return nil, err
	}
	vid := uint16(vid64)

	// build CipherSuites
	var suites []uint16
	for _, c := range ciphers {
		cid, err := strconv.ParseUint(c, 10, 16)
		if err != nil {
			return nil, err
		}
		suites = append(suites, uint16(cid))
	}

	return &tls.ClientHelloSpec{
		TLSVersMin:         vid,
		TLSVersMax:         vid,
		CipherSuites:       suites,
		CompressionMethods: []byte{0},
		Extensions:         exts,
		GetSessionID:       sha256.Sum256,
	}, nil
}

func urlToHost(target *url.URL) *url.URL {
	if !strings.Contains(target.Host, ":") {
		if target.Scheme == "http" {
			target.Host = target.Host + ":80"
		} else if target.Scheme == "https" {
			target.Host = target.Host + ":443"
		}
	}
	return target
}


// JA3Response is the struct
type JA3Response struct {
	JA3Hash   string `json:"ja3_hash"`
	JA3       string `json:"ja3"`
	UserAgent string `json:"User-Agent"`
}

// JA3Client contains is similar to http.Client
type JA3Client struct {
	*http.Client

	Config  *tls.Config
	Browser Browser
}

// New creates a JA3Client based on a Browser struct
func New(b Browser) (*JA3Client, error) {
	client, err := NewWithString(b.JA3)
	if err != nil {
		return nil, err
	}
	client.Browser = b
	return client, nil
}

// NewWithString creates a JA3 client with the specified JA3 string
func NewWithString(ja3 string) (*JA3Client, error) {
	tr, err := NewTransport(ja3)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Transport: tr}

	return &JA3Client{
		client,
		&tls.Config{},
		Browser{JA3: ja3},
	}, nil
}

// Do sends an HTTP request and returns an HTTP response, following policy
// (such as redirects, cookies, auth) as configured on the client.
func (c *JA3Client) Do(req *http.Request) (*http.Response, error) {
	if _, ok := req.Header["User-Agent"]; !ok && c.Browser.UserAgent != "" {
		req.Header.Set("User-Agent", c.Browser.UserAgent)
	}
	return c.Client.Do(req)
}

// Get issues a GET to the specified URL.
func (c *JA3Client) Get(targetURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post issues a POST to the specified URL.
func (c *JA3Client) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// Head issues a HEAD to the specified URL.
func (c *JA3Client) Head(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// PostForm issues a POST to the specified URL,
// with data's keys and values URL-encoded as the request body.
func (c *JA3Client) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}



func main() {


	httpClient,err := New(SafariAuto)
	if err != nil{
		fmt.Println(err)
		panic(err)
	}
	// realUrl := "https://ja3er.com/json"
	
	/* First fetch the JA3 Fingerprint */
	resp, err := httpClient.Get("https://www.thegreatcodeadventure.com/mocking-http-requests-in-golang/")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// unmarshal the response
	var ja3Response JA3Response
	err = json.Unmarshal(body, &ja3Response)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}


	var out bytes.Buffer
	err = json.Indent(&out, body, "", "\t")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println(out.String())
}
