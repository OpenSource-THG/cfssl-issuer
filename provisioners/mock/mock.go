package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/cloudflare/cfssl/api"
)

func New() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cfssl/sign", mockSign)
	return httptest.NewTLSServer(mux)
}

func mockSign(w http.ResponseWriter, r *http.Request) {
	cert, err := os.ReadFile("testdata/client.pem")
	if err != nil {
		http.Error(w, fmt.Errorf("fail to load cert: %v", err).Error(), http.StatusInternalServerError)
		return
	}

	resp := api.Response{
		Success: true,
		Result: map[string]string{
			"certificate": string(cert),
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}
