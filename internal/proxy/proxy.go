package proxy

import (
	"fmt"
	"net/http"

	"github.com/spf13/viper"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
}

func StartProxy() error {
	http.HandleFunc("/", handleRequest)

	return http.ListenAndServe(viper.GetString("proxy_address"), nil)
}
