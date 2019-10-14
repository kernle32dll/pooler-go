package main

import (
	"github.com/kernle32dll/pooler-go"

	"encoding/json"
	"log"
	"net/http"
	"time"
)

// User is a just sample struct for showcasing.
type User struct {
	Name string
	Time time.Time
}

const poolerKey = iota

func main() {
	router := http.NewServeMux()

	middleware := pooler.NewMiddleware(poolerKey, func() interface{} {
		log.Println("New object created")
		return &User{}
	})

	router.Handle("/user", middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := pooler.Get(r.Context(), poolerKey).(*User)

			user.Name = "Björn Gerdau"
			user.Time = time.Now()

			decoder := json.NewEncoder(w)
			if err := decoder.Encode(user); err != nil {
				panic(err)
			}
		}),
	))

	log.Fatal(http.ListenAndServe(":8080", router))
}
