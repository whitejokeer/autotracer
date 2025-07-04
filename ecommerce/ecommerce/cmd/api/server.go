package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/example/ecommerce/internal/inventory"
	"github.com/example/ecommerce/internal/notification"
	"github.com/example/ecommerce/internal/order"
	"github.com/example/ecommerce/internal/payment"
	"github.com/example/ecommerce/internal/user"
)

type Server struct {
	router      *mux.Router
	orderSvc    *order.Service
	paymentSvc  *payment.Service
	userSvc     *user.Service
}

func NewServer() *Server {
	inventorySvc := inventory.NewService()
	notificationSvc := notification.NewService()
	paymentGateway := payment.NewGateway()
	
	paymentSvc := payment.NewService(paymentGateway, inventorySvc, notificationSvc)
	orderSvc := order.NewService(paymentSvc, inventorySvc)
	userSvc := user.NewService()

	s := &Server{
		router:     mux.NewRouter(),
		orderSvc:   orderSvc,
		paymentSvc: paymentSvc,
		userSvc:    userSvc,
	}

	s.setupRoutes()
	return s
}

func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/api/orders", s.handleCreateOrder).Methods("POST")
	s.router.HandleFunc("/api/orders/{id}", s.handleGetOrder).Methods("GET")
	s.router.HandleFunc("/api/users/register", s.handleUserRegistration).Methods("POST")
	s.router.HandleFunc("/api/users/login", s.handleUserLogin).Methods("POST")
	s.router.HandleFunc("/api/users/{id}/profile", s.handleUpdateProfile).Methods("PUT")
}

func (s *Server) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req order.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	orderResp, err := s.orderSvc.CreateOrder(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orderResp)
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	order, err := s.orderSvc.GetOrder(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (s *Server) handleUserRegistration(w http.ResponseWriter, r *http.Request) {
	var req user.RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userResp, err := s.userSvc.RegisterUser(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResp)
}

func (s *Server) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	var req user.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := s.userSvc.Login(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var req user.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.userSvc.UpdateProfile(r.Context(), userID, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}