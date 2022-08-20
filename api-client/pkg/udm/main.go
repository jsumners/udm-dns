package udm

import (
	"crypto/tls"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type UdmConfig struct {
	Address string `mapstructure:"address"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Site string `mapstructure:"site"`
}

type IUdmClient interface {
	GetActiveClients() []NetworkClient
	GetConfiguredClients() []NetworkClient
}

type UdmClient struct {
	config UdmConfig
	resty resty.Client
}

// Represents the `meta` property in an
// UDM API response.
type ResponseMeta struct {
	// Typically "ok".
	Code string `json:"rc"`
	Message string `json:"msg"`

	// TODO: I believe a `count` can also be returned
	// to indicate pagination. But I don't have enough
	// data to work with to determine how pagination
	// would work.
}

type NetworkClient struct {
	Hostname string `json:"hostname"`
	FixedIpAddress string `json:"fixed_ip"`
	IpAddress string `json:"ip"`
	MacAddress string `json:"mac"`
	Name string `json:"name"`
}

// Represents the response sent when querying the
// UDM API for a list of network clients.
type NetworkClientsResponse struct {
	Meta ResponseMeta `json:"meta"`
	Data []NetworkClient `json:"data"`
}

// Instantiate a [udm.UdmClient] instance and return it.
// The instance will be authenticated with the remote UDM server
// and ready to issue requests. If there was a problem logging in,
// a `panic` will be issued.
func New(udmConfig UdmConfig) *UdmClient {
	client := &UdmClient{
		config: udmConfig,
	}

	client.resty = *resty.New()
	client.resty.SetBaseURL(fmt.Sprintf("https://%s", udmConfig.Address))
	client.resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	client.login()

	return client
}

func (udm *UdmClient) login() {
	payload := fmt.Sprintf(
		`{"username":"%s","password":"%s"}`,
		udm.config.Username,
		udm.config.Password,
	)

	resp, err := udm.resty.R().
		SetHeader("content-type", "application/json").
		SetBody(payload).
		Post("/api/auth/login")

	if err != nil {
		panic(err)
	}
	if resp.StatusCode() >= 400 {
		panic(fmt.Errorf("login error: %s", resp.Status()))
	}
}

func (udm *UdmClient) GetActiveClients() []NetworkClient {
	return udm.getNeworkClients(
		fmt.Sprintf("/proxy/network/api/s/%s/stat/sta", udm.config.Site),
	)
}

func (udm *UdmClient) GetConfiguredClients() []NetworkClient {
	return udm.getNeworkClients(
		fmt.Sprintf("/proxy/network/api/s/%s/list/user", udm.config.Site),
	)
}

func (udm *UdmClient) getNeworkClients(path string) []NetworkClient {
	resp, err := udm.resty.R().Get(path)
	if err != nil {
		panic(err)
	}

	parsed := NetworkClientsResponse{}
	unmarshalError := json.Unmarshal(resp.Body(), &parsed)
	if unmarshalError != nil {
		panic(unmarshalError)
	}

	if parsed.Meta.Code == "error" {
		panic(fmt.Errorf("api error: %s", parsed.Meta.Message))
	}

	return parsed.Data
}
