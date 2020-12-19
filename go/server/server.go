package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/go-redis/redis"
)

type IntakeData struct {
	User		int64 		`json:"user"`
	Event		int64 		`json:"event"`
	Req_Id 		string 		`json:"req_id"`
	Position 	[]float64	`json:"position"`
}

var redisPool *redis.ClusterClient

func initialize(hostnames string) *redis.ClusterClient {
	addr := strings.Split(hostnames, ",")
	c := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addr,
		PoolSize: 10,
	})
	if err := c.Ping().Err(); err != nil {
		panic("Unable to connect to redis " + err.Error())
	}

	return c
}

func intake(w http.ResponseWriter, req *http.Request) {
	
	b, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var msg IntakeData
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Println(msg.Req_Id)

	pubsub := redisPool.Subscribe(fmt.Sprintf("%d-%d-%s-reply", msg.Event, msg.User, msg.Req_Id))

	done := make(chan struct{})

	go func() {
		defer close(done)

		redisMsg, redisErr := pubsub.ReceiveMessage()
		if redisErr != nil {
			panic(redisErr)
		}

		fmt.Println(redisMsg.Channel, redisMsg.Payload)

		redisErr = pubsub.Close()
		if redisErr != nil {
			panic(redisErr)
		}

		w.Header().Set("content-type", "application/json")
		w.Write([]byte(redisMsg.Payload))

	}()

	fmt.Printf("%d-%d-pos\n", msg.Event, msg.User)
	err = redisPool.Publish(fmt.Sprintf("%d-%d-pos", msg.Event, msg.User), string(b)).Err()
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <- done:
			return
		}
	}

}

func main() {

	redisPool = initialize("127.0.0.1:7000,127.0.0.1:7001,127.0.0.1:7002")

	http.HandleFunc("/intake", intake)

	port := ":" + os.Getenv("INGESTION_PORT")
	fmt.Println("Listening on port " + port)

	http.ListenAndServe(port, nil)
}
