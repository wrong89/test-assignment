package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	httpHandlers *HTTPHandlers
}

func NewHTTPServer(httpHandlers *HTTPHandlers) *HTTPServer {
	return &HTTPServer{
		httpHandlers: httpHandlers,
	}
}

func (s *HTTPServer) StartServer(address string) error {
	router := mux.NewRouter()

	router.Use(AuthMiddleware)

	router.Path("/v1/user").Methods("POST").HandlerFunc(s.httpHandlers.HandleCreateUser)
	router.Path("/v1/withdrawals").Methods("POST").HandlerFunc(s.httpHandlers.HandleWithdrawal)
	router.Path("/v1/withdrawals/{id}").Methods("GET").HandlerFunc(s.httpHandlers.HandleGetWithdrawal)

	return http.ListenAndServe(address, router)
}
