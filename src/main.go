package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter,r *http.Request)
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	hanleErr(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func newLoadBalancer(port string, server []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         server,
	}
}

func hanleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *simpleServer) Address() string {
	return s.addr
}

func (s *simpleServer)IsAlive() bool{
	return true;
}
func (s *simpleServer)Serve(rw http.ResponseWriter, r *http.Request) {

	s.proxy.ServeHTTP(rw, r)
}

func (lb *LoadBalancer) getNxtAvaibaleServer() Server{
	server := lb.servers[lb.roundRobinCount% len(lb.servers)]
	for !server.IsAlive(){
		lb.roundRobinCount++;
		server = lb.servers[lb.roundRobinCount% len(lb.servers)]
	}
	lb.roundRobinCount++;
	return server
}

func (lb *LoadBalancer)serveProxy(rw http.ResponseWriter, r *http.Request){
	targetServer:= lb.getNxtAvaibaleServer()
	fmt.Printf("forwarding request to adress %q\n", targetServer.Address())
	targetServer.Serve(rw, r)
}

func main(){
	servers := []Server{
		newSimpleServer("http://www.bing.com"),
		newSimpleServer("http://www.facebook.com"),
		newSimpleServer("http://www.x.com"),
	}
	lb := newLoadBalancer("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, r *http.Request){
		lb.serveProxy(rw, r)
	}

	http.HandleFunc("/", handleRedirect);

	fmt.Printf("Serving at port: %s\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
