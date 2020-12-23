package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const (
	redisComputeStartUrl = "http://localhost:2930/race"
	sprintFinishStartUrl = "http://localhost:2929/race"
	redisComputeDestroyUrl = "http://localhost:2930/destroyRace"
	sprintFinishDestroyUrl = "http://localhost:2929/destroyRace"
)

type SprintStatus struct {
	Num 		int64 	`json:"num"`
	Started 	bool 	`json:"started"`
}

type Racer struct {
	Uid 	int64 	`json:"uid"`
	Rid 	int64 	`json:"rid"`
}

type RaceMetadata struct {
	Rid 		int64 		`json:"rid"`
	Racers		[]Racer 	`json:"racers"`
	NumSprints 	int 		`json:"numSprints"`
}

type IntakeData struct {
	User		int64 		`json:"user"`
	Event		int64 		`json:"event"`
	Req_Id 		string 		`json:"req_id"`
	Position 	[]float64	`json:"position"`
}

type LeaderboardEntry struct {
	Uid			string 		`json:"uid"`
	Chkpt		float64 	`json:"chkpt"`
}

type LastSprintPlace struct {
	Place 		int64 		`json:"place"`
	Points 		int64 		`json:"points"`
}

type RedisData struct {
	Uid				string 		`json:"uid"`
	Checkpoint		float64 	`json:"checkpoint"`
	InSprintZone 	bool 		`json:"inSprintZone"`
	LeaderBoard 	[]LeaderboardEntry `json:"leaderboard"`
	LastSprint 		LastSprintPlace `json:"last_sprint_place"`
}

type ReturnData struct {
	Data 	RedisData 	`json:"data"`
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
		fmt.Println(err.Error());
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Println(msg.Position)

	if msg.Position[0] > 90 {
		http.Error(w, "invalid position", 401)
		return
	}

	pubsub := redisPool.Subscribe(fmt.Sprintf("%d-%d-%s-reply", msg.Event, msg.User, msg.Req_Id))
	pubsubChan := pubsub.Channel()

	done := make(chan struct{})

	go func() {
		defer close(done)

		timer := time.NewTimer(time.Millisecond * 500);
		var redisMsg *redis.Message

		for {
			select {

			case redisMsg, _ = <-pubsubChan:
				fmt.Println(redisMsg)
				break;

			case <-timer.C:
				http.Error(w, "Server timeout", 408)
				return
			}

			break
		}

		redisErr := pubsub.Close()
		if redisErr != nil {
			panic(redisErr)
		}

		w.Header().Set("content-type", "application/json")
		w.Write([]byte(redisMsg.Payload))

	}()

	// fmt.Printf("%d-%d-pos\n", msg.Event, msg.User)
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

func startRace(w http.ResponseWriter, req *http.Request) {

	b, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var msg RaceMetadata
	err = json.Unmarshal(b, &msg)
	if err != nil {
		fmt.Println(err.Error());
		http.Error(w, err.Error(), 500)
		return
	}

	sprintSet := SprintStatus{
		Num: 0,
		Started: false,
	}

	sprintSetBytes, _ := json.Marshal(sprintSet)

	redisPool.Del(fmt.Sprintf("%d-leaderboard", msg.Rid))

	for _, r := range msg.Racers {

		redisPool.Set(fmt.Sprintf("%d-%d-chkpt", msg.Rid, r.Uid), 0, 0)
		redisPool.Set(fmt.Sprintf("%d-%d-pts", msg.Rid, r.Uid), 0, 0)
		redisPool.Set(fmt.Sprintf("%d-%d-sprint_num", msg.Rid, r.Uid), string(sprintSetBytes), 0)
		redisPool.Del(fmt.Sprintf("%d-%d-pos", msg.Rid, r.Uid))
		redisPool.Set(fmt.Sprintf("%d-numRacers", msg.Rid), len(msg.Racers), 0)
		redisPool.Set(fmt.Sprintf("%d-finished", msg.Rid), 0, 0)

		for i := 0; i < msg.NumSprints; i++ {
			redisPool.Del(fmt.Sprintf("%d-%d-%d", msg.Rid, r.Uid, i))
			redisPool.Del(fmt.Sprintf("%d-%d-%d-place", msg.Rid, r.Uid, i))
		}

	}

	for i := 0; i < msg.NumSprints; i++ {
		redisPool.Del(fmt.Sprintf("%d-%d-sprint_finish", msg.Rid, i))
	}

	_, err = http.Post(redisComputeStartUrl, "application/json", bytes.NewBuffer(b))
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	_, err = http.Post(sprintFinishStartUrl, "application/json", bytes.NewBuffer(b))
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.WriteHeader(200)

}

func finish(w http.ResponseWriter, req *http.Request) {

	b, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var racer Racer
	err = json.Unmarshal(b, &racer)
	if err != nil {
		fmt.Println(err.Error());
		http.Error(w, err.Error(), 500)
		return
	}

	numRacers, _ := redisPool.Get(fmt.Sprintf("%d-numRacers", racer.Rid)).Result()
	numFinished, _ := redisPool.Incr(fmt.Sprintf("%d-finished", racer.Rid)).Result()

	numRacersInt, _ := strconv.Atoi(numRacers)

	if numRacersInt == int(numFinished) {
		destroyRace(racer)
	}

}


func destroyRace(racer Racer) {

	racerBytes, _ := json.Marshal(racer)

	_, err := http.Post(redisComputeDestroyUrl, "application/json", bytes.NewBuffer(racerBytes))
	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = http.Post(sprintFinishDestroyUrl, "application/json", bytes.NewBuffer(racerBytes))
	if err != nil {
		fmt.Println(err.Error())
	}

}

func main() {

	redisPool = initialize("127.0.0.1:7000,127.0.0.1:7001,127.0.0.1:7002")

	http.HandleFunc("/intake", intake)
	http.HandleFunc("/startRace", startRace)
	http.HandleFunc("/finish", finish)

	port := ":" + os.Getenv("INGESTION_PORT")
	fmt.Println("Listening on port " + port)

	if os.Getenv("INGESTION_PORT") == "443" {
		http.ListenAndServeTLS(port, os.Getenv("SSL_CERT_FILENAME"), os.Getenv("SSL_KEY_FILENAME"), nil)
	} else {
		http.ListenAndServe(port, nil)
	}
	
}
