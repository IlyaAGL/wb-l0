package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/agl/wbtech/internal/application/interfaces"
	"github.com/agl/wbtech/pkg/logger"
)

type OrderController struct {
	port    string
	service interfaces.OrderService
}

func NewOrderController(service interfaces.OrderService) *OrderController {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &OrderController{
		port:    port,
		service: service,
	}
}

func (oc *OrderController) StartServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/orders/", oc.getOrderByID)

	logger.Log.Info("Starting server", "port", oc.port)

	if err := http.ListenAndServe(fmt.Sprintf(":%v", oc.port), mux); err != nil {
		panic(err)
	}
}

func (oc *OrderController) getOrderByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	path := r.URL.Path
	prefix := "/orders/"
	if len(path) <= len(prefix) || path[:len(prefix)] != prefix {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	id := path[len(prefix):]
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	order, err := oc.service.GetOrderByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
