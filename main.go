package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	router "ticketing-service/api"
	"time"

	//"ticketing-service/data"
	//utils "ticketing-service/util"

	"github.com/gorilla/handlers"
	//data "ticketing-service/model"
	//"ticketing-service/service"
)

var listenAddr string

func main() {
	//listenAddr = "8080"
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "server liste address")
	flag.Parse()
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// http://localhost:8080/swaggerui/#/ to access the service

	server := newWebserver()
	go gracefullShutdown(server, quit, done)
	fmt.Printf("Server is ready to handle requests at %s", listenAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Could not listen on %s: %v\n", listenAddr, err)
	}
	<-done
	fmt.Printf("Server stopped")
}

func gracefullShutdown(server *http.Server, quit <-chan os.Signal, done chan<- bool) {
	sig := <-quit
	fmt.Printf("Server is shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Could not gracefully shutdown the server: %v\n", err)
	}
	close(done)
}

func processRequestURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		fmt.Printf("URL Path :%s", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func newWebserver() *http.Server {

	router := router.NewRouter()
	headers := handlers.AllowedHeaders([]string{"Host", "Origin", "Connection", "Upgrade", "Sec-WebSocket-Key", "Sec-WebSocket-Version", "X-Requested-With", "Content-Type", "Authorization", "Accept", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "access-control-allow-origin", "access-control-allow-headers"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS", "DELETE"})
	origins := handlers.AllowedOrigins([]string{"*"})
	credentials := handlers.AllowCredentials()
	_, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	var fs http.Handler
	fs = http.FileServer(http.Dir("./swaggerui/"))
	router.PathPrefix("/swaggerui").Handler(http.StripPrefix("/swaggerui", fs))

	//var handler http.Handler
	// configs := utils.NewConfigurations()
	// validator := data.NewValidation()
	// authService := service.NewAuthService(configs)
	// uh := router.NewAuthHandler(validator, authService)

	return &http.Server{
		Addr:         listenAddr,
		Handler:      handlers.CORS(headers, methods, origins, credentials)(processRequestURL(router)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}
