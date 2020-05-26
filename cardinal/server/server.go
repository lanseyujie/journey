package server

import (
    "crypto/tls"
    "fmt"
    "net"
    "net/http"
    "time"
)

type Server struct {
    *http.Server
    listener  net.Listener
    errorChan chan error
}

var cluster = make([]*Server, 0, 2)

// NewServer
func NewServer(addr string, handler http.Handler) *Server {
    srv := &Server{
        Server: &http.Server{
            Addr:              addr,
            Handler:           handler,
            TLSConfig:         nil,
            ReadTimeout:       10 * time.Second,
            ReadHeaderTimeout: 5 * time.Second,
            WriteTimeout:      15 * time.Second,
            IdleTimeout:       120 * time.Second,
            MaxHeaderBytes:    0,
            TLSNextProto:      nil,
            ConnState:         nil,
            ErrorLog:          nil,
            BaseContext:       nil,
            ConnContext:       nil,
        },
        errorChan: make(chan error),
    }

    cluster = append(cluster, srv)

    return srv
}

// ListenAndServe
func (srv *Server) ListenAndServe() (err error) {
    addr := srv.Addr
    if addr == "" {
        addr = ":http"
    }

    return srv.Serve()
}

// ListenAndServeTLS
func (srv *Server) ListenAndServeTLS(certFile, keyFile string) (err error) {
    addr := srv.Addr
    if addr == "" {
        addr = ":https"
    }

    if srv.TLSConfig == nil {
        srv.TLSConfig = &tls.Config{}
    }

    if srv.TLSConfig.NextProtos == nil {
        srv.TLSConfig.PreferServerCipherSuites = true
        srv.TLSConfig.NextProtos = []string{"h2", "http/1.1"}
        // srv.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
    }

    srv.TLSConfig.Certificates = make([]tls.Certificate, 1)
    srv.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return
    }

    return srv.Serve()
}

// getListener
func (srv *Server) getListener() (err error) {
    if srv.TLSConfig == nil {
        srv.listener, err = net.Listen("tcp", srv.Addr)
    } else {
        srv.listener, err = tls.Listen("tcp", srv.Addr, srv.TLSConfig)
    }

    if err != nil {
        return fmt.Errorf("net.Listen error: %v", err)
    }

    return
}

// Serve
func (srv *Server) Serve() (err error) {
    err = srv.getListener()
    if err != nil {
        return
    }

    err = srv.Server.Serve(srv.listener)
    if err != nil && err != http.ErrServerClosed {
        return
    }

    return nil
}
