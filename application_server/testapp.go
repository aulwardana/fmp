package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"encoding/json"

	"gopkg.in/mgo.v2"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	hosts      = "localhost:27017"
	database   = "kijang"
	username   = "aulwardana"
	password   = "rahasia"
	collection = "sensor"
)

type temp struct {
	Code        string  `json:"Code"`
	Temperature float32 `json:"Temperature"`
	Humidity    float32 `json:"Humidity"`
}

var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	info := &mgo.DialInfo{
		Addrs:    []string{hosts},
		Timeout:  60 * time.Second,
		Database: database,
		Username: username,
		Password: password,
	}

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())

	sensing := fmt.Sprintf("%s", msg.Payload())
	res := temp{}
	json.Unmarshal([]byte(sensing), &res)

	c := session.DB(database).C(collection)
	err = c.Insert(res)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	runtime.GOMAXPROCS(1)

	var wg sync.WaitGroup
	wg.Add(2)

	opts := MQTT.NewClientOptions().AddBroker("tcp://192.168.2.27:1883")
	opts.SetClientID("go-simple")
	opts.SetDefaultPublishHandler(f)

	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	go func() {
		defer wg.Done()

		if token := c.Subscribe("test", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
	}()

	wg.Wait()
}
