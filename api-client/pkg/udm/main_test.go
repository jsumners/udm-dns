package udm

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testServer struct {
	instance *httptest.Server
	address string
}

type httpAssertionCallback func(req *http.Request, body string) (int, string)
type testServerParams struct {
	t *testing.T
	assertionsCallback httpAssertionCallback
}

// Creates an HTTP server to test log delivery payloads by applying a set of
// assertions through the assertCB function.
func createHTTPServer(params *testServerParams) testServer {
	httpServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			params.t.Fatal(err)
		}

		statusCode, responseBody := params.assertionsCallback(r, string(body))

		if len(responseBody) > 0 {
			w.Header().Add("content-type", "application/json")
			w.WriteHeader(statusCode)
			w.Write([]byte(responseBody))
		} else {
			w.WriteHeader(statusCode)
		}
	}))

	server := testServer{
		instance: httpServer,
		address: httpServer.Listener.Addr().String(),
	}

	return server
}

func TestNew(t *testing.T) {
	serverParams := &testServerParams{
		t: t,
		assertionsCallback: func(req *http.Request, body string) (int, string) {
			assert.Fail(t, "stubbed assertions callback invoked")
			return 0, ""
		},
	}

	server := createHTTPServer(serverParams)
	defer server.instance.Close()

	udmConfig := UdmConfig{
		Address: server.address,
		Username: "test",
		Password: "test",
		Site: "default",
	}

	t.Run("bad credentials", func (t *testing.T) {
		serverParams.assertionsCallback = func(req *http.Request, body string) (int, string) {
			assert.Equal(t, "/api/auth/login", req.URL.Path)
			return http.StatusForbidden, `{
				"code": "AUTHENTICATION_FAILED_INVALID_CREDENTIALS",
  			"message": "Invalid username or password"
			}`
		}
		assert.Panics(t, func() { New(udmConfig) })
	})

	t.Run("good credentials", func (t *testing.T) {
		serverParams.assertionsCallback = func(req *http.Request, body string) (int, string) {
			assert.Equal(t, "/api/auth/login", req.URL.Path)
			assert.Equal(t, body, `{"username":"test","password":"test"}`)
			return http.StatusOK, ""
		}

		client := New(udmConfig)
		assert.NotNil(t, client)
	})

	t.Run("server error", func (t *testing.T) {
		serverParams.assertionsCallback = func(req *http.Request, body string) (int, string) {
			assert.Equal(t, "/api/auth/login", req.URL.Path)
			return http.StatusInternalServerError, `{"bad":"json"`
		}
		assert.Panics(t, func() { New(udmConfig) })
	})
}

func TestGetConfiguredClients(t *testing.T) {
	serverParams := &testServerParams{
		t: t,
		assertionsCallback: func(req *http.Request, body string) (int, string) {
			// First request is login request so send OK.
			return http.StatusOK, ""
		},
	}

	server := createHTTPServer(serverParams)
	defer server.instance.Close()

	udmConfig := UdmConfig{
		Address: server.address,
		Username: "test",
		Password: "test",
		Site: "default",
	}
	udmClient := New(udmConfig)

	t.Run("panics for bad json", func(t *testing.T) {
		serverParams.assertionsCallback = func(req *http.Request, body string) (int, string) {
			assert.Equal(t, "/proxy/network/api/s/default/stat/sta", req.URL.Path)
			return http.StatusOK, `{"bad":"json"`
		}
		assert.Panics(t, func() { udmClient.GetActiveClients() })
	})

	t.Run("panics for metadata error", func(t *testing.T) {
		serverParams.assertionsCallback = func(req *http.Request, body string) (int, string) {
			assert.Equal(t, "/proxy/network/api/s/default/stat/sta", req.URL.Path)
			return http.StatusOK, `{
				"meta": {
					"rc": "error",
					"msg": "foo bar"
				}
			}`
		}
		assert.PanicsWithError(t, "api error: foo bar", func() { udmClient.GetActiveClients() })
	})

	t.Run("returns network clients", func(t *testing.T) {
		serverParams.assertionsCallback = func(req *http.Request, body string) (int, string) {
			assert.Equal(t, "/proxy/network/api/s/default/stat/sta", req.URL.Path)
			return http.StatusOK, `{
				"meta": { "rc": "ok" },
				"data": [
					{
						"hostname": "foo",
						"fixed_ip": "10.0.0.2",
						"ip": "10.0.0.2",
						"mac": "a_mac",
						"name": "foo"
					}
				]
			}`
		}

		networkClients := udmClient.GetActiveClients()
		assert.Equal(t, 1, len(networkClients))
	})
}

func TestGetActiveClients(t *testing.T) {
	serverParams := &testServerParams{
		t: t,
		assertionsCallback: func(req *http.Request, body string) (int, string) {
			// First request is login request so send OK.
			return http.StatusOK, ""
		},
	}

	server := createHTTPServer(serverParams)
	defer server.instance.Close()

	udmConfig := UdmConfig{
		Address: server.address,
		Username: "test",
		Password: "test",
		Site: "default",
	}
	udmClient := New(udmConfig)

	t.Run("panics for bad json", func(t *testing.T) {
		serverParams.assertionsCallback = func(req *http.Request, body string) (int, string) {
			assert.Equal(t, "/proxy/network/api/s/default/list/user", req.URL.Path)
			return http.StatusOK, `{"bad":"json"`
		}
		assert.Panics(t, func() { udmClient.GetConfiguredClients() })
	})

	t.Run("panics for metadata error", func(t *testing.T) {
		serverParams.assertionsCallback = func(req *http.Request, body string) (int, string) {
			assert.Equal(t, "/proxy/network/api/s/default/list/user", req.URL.Path)
			return http.StatusOK, `{
				"meta": {
					"rc": "error",
					"msg": "foo bar"
				}
			}`
		}
		assert.PanicsWithError(t, "api error: foo bar", func() { udmClient.GetConfiguredClients() })
	})

	t.Run("returns network clients", func(t *testing.T) {
		serverParams.assertionsCallback = func(req *http.Request, body string) (int, string) {
			assert.Equal(t, "/proxy/network/api/s/default/list/user", req.URL.Path)
			return http.StatusOK, `{
				"meta": { "rc": "ok" },
				"data": [
					{
						"hostname": "foo",
						"fixed_ip": "10.0.0.2",
						"ip": "10.0.0.2",
						"mac": "a_mac",
						"name": "foo"
					}
				]
			}`
		}

		networkClients := udmClient.GetConfiguredClients()
		assert.Equal(t, 1, len(networkClients))
	})
}
