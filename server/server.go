package server

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/spf13/viper"
	"github.com/t94j0/ja3-server/net/http"
	"github.com/t94j0/satellite/path"
)

func Start(paths *path.Paths, cfg *viper.Viper) error {
	//serverRoot := cfg.GetString("server_root")
	keyPath := cfg.GetString("ssl.key")
	certPath := cfg.GetString("ssl.cert")
	port := cfg.GetString("listen")
	//managementIP := cfg.GetString("management.ip")
	//managementPath := cfg.GetString("management.path")

	//mux := http.NewServeMux()
	//if managementPath != "" {
	//	mgmt := handler.NewManagementHandler(paths, serverRoot, managementIP)
	//	mux.Handle(managementPath, mgmt)
	//}
	//mux.Handle("/", handler.NewRoot(paths, cfg))
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Println("we out here")
	})

	server := &http.Server{Addr: port, Handler: handler}
	ln, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer ln.Close()
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}
	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
	tlsListener := tls.NewListener(ln, &tlsConfig)
	return server.Serve(tlsListener)
}
