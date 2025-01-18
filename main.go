package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type OrderResponse struct {
	OrderId string `json:"order_id"`
	Status  string `json:"status"`
}

func main() {
	router := httprouter.New()

	router.GET("/events", EventHandler)

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func EventHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Add("Access-Control-Allow-Origin", "*")

	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	status := []string{"Paid", "Unpaid"}

breakLabel:
	for {
		select {
		case <-ticker.C:
			log.Println("Write data")

			response := OrderResponse{
				OrderId: uuid.New().String(),
				Status:  status[rand.Intn(len(status))],
			}

			WriteToResponseBody(writer, response)

			flusher.Flush()

		case <-request.Context().Done():
			log.Println("Client disconnected")
			break breakLabel
		}
	}
}

func WriteToResponseBody(writer http.ResponseWriter, response interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Println("Error encoding response:", err)
		return
	}

	_, err = writer.Write([]byte("data: "))
	if err != nil {
		log.Println("Error writing data prefix:", err)
		return
	}

	_, err = writer.Write(data)
	if err != nil {
		log.Println("Error writing JSON data:", err)
		return
	}

	_, err = writer.Write([]byte("\n\n"))
	if err != nil {
		log.Println("Error writing event end:", err)
	}
}
