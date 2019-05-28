package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/gorilla/mux"
)

var dataFile = path.Join(".", "data", "proverbs.json")
var hostAddress string
var hostPort string

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}
	hostAddress = getEnvOrDefault("HOST_ADDRESS", "127.0.0.1")
	hostPort = getEnvOrDefault("HOST_PORT", "80")
}

func main() {
	proverbs, err := loadProverbs(dataFile)
	if err != nil {
		log.Fatalln(err)
	}

	h := newHandler(proverbs)
	r := newRouter(h)

	var sigChan = make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		log.Printf("Signal received: %+v.", <-sigChan)
		log.Println("Saving proverbs...")
		// cleanup code goes here - shutdown nicely
		if err := saveProverbs(dataFile, h.proverbs); err != nil {
			log.Printf("Something went wrong: %s.", err)
		}
		log.Println("Bye.")
		os.Exit(0)
	}()

	hostAndPort := hostAddress + ":" + hostPort
	log.Println(fmt.Sprintf("API Server starting on %s ...", hostAndPort))
	log.Fatalln(http.ListenAndServe(hostAndPort, r))
}

func newRouter(h *handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/proverbs", h.createProverb).Methods("POST")
	r.HandleFunc("/proverbs", h.getProverbs).Methods("GET")
	r.HandleFunc("/proverbs/{id:[0-9]+}", h.getProverb).Methods("GET")
	r.HandleFunc("/proverbs/{id:[0-9]+}", h.updateProverb).Methods("PUT")
	r.HandleFunc("/proverbs/{id:[0-9]+}", h.deleteProverb).Methods("DELETE")
	return r
}

func loadProverbs(dataFile string) ([]Proverb, error) {
	file, err := os.Open(dataFile)
	if err != nil {
		return nil, err
	}

	var proverbs []Proverb
	if err := json.NewDecoder(file).Decode(&proverbs); err != nil {
		return nil, err
	}

	return proverbs, nil
}

func saveProverbs(dataFile string, proverbs []Proverb) error {
	file, err := os.Create(dataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(proverbs)
}

func getEnvOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
