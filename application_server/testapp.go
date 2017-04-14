package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"gopkg.in/mgo.v2"
	//	"gopkg.in/mgo.v2/bson"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	hosts      = "localhost:27017"
	database   = "kijang"
	username   = "aulwardana"
	password   = "rahasia"
	collection = "sensor"

	host    = "localhost"
	port    = 5432
	user    = "postgres"
	pwd     = "postgres"
	dbname  = "postgres"
	sslmode = "disable"

	hostMQTT = "tcp://localhost:1883"

	webappPort = ":8000"

	DefLatitude  = "-6.890546"
	DefLongitude = "107.609505"
)

var templateHome = template.Must(template.ParseFiles("index.html"))
var templateInsert = template.Must(template.ParseFiles("insert.html"))
var templateMonitoring = template.Must(template.ParseFiles("monitor.html"))

type Sensor struct {
	Code     string    `json:"Code"`
	Distance float32   `json:"Distance"`
	Date     time.Time `json:"Date"`
}

type Place struct {
	Uid       int
	Name      string
	Type      string
	Location  string
	State     string
	Kawasan   string
	Latitude  string
	Longitude string
	Code      string
}

type DefaultPlace struct {
	Latitude  string
	Longitude string
}

func FloatToString(input_num float64) string {
	return strconv.FormatFloat(input_num, 'f', 2, 64)
}

func HomeTemplate(w http.ResponseWriter, r *http.Request) {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, pwd, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	Places := []Place{}

	rows, err := db.Query("SELECT * FROM public.place")
	if err != nil {
		http.Error(w, "Select error", 500)
		return
	}

	for rows.Next() {
		c := Place{}
		err := rows.Scan(&c.Uid, &c.Name, &c.Type, &c.Location, &c.State, &c.Kawasan, &c.Latitude, &c.Longitude, &c.Code)
		if err != nil {
			http.Error(w, "Decompose error", 500)
			return
		}

		Places = append(Places, c)
	}

	templateHome.ExecuteTemplate(w, "index.html", Places)
}

func InsertTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		DefaultPlaces := DefaultPlace{DefLatitude, DefLongitude}
		templateInsert.ExecuteTemplate(w, "insert.html", DefaultPlaces)
		return
	}

	nameJ := r.FormValue("name")
	typeJ := r.FormValue("type")
	locationJ := r.FormValue("location")
	stateJ := r.FormValue("state")
	kawasanJ := r.FormValue("kawasan")
	latitudeJ := r.FormValue("lat")
	longitudeJ := r.FormValue("long")
	codeJ := r.FormValue("code")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, pwd, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	var lastInsertId int
	err = db.QueryRow("INSERT INTO public.place(name, type, location, state, kawasan, latitude, longitude, code) VALUES($1,$2,$3,$4,$5,$6,$7,$8) returning uid;", nameJ, typeJ, locationJ, stateJ, kawasanJ, latitudeJ, longitudeJ, codeJ).Scan(&lastInsertId)
	log.Println("last inserted id =", lastInsertId)

	if err != nil {
		http.Error(w, "Insert error", 500)
		return
	} else {
		http.Redirect(w, r, "/", 301)
		return
	}
}

func MonitoringTemplate(w http.ResponseWriter, r *http.Request) {
	DefaultPlaces := DefaultPlace{DefLatitude, DefLongitude}
	templateMonitoring.ExecuteTemplate(w, "monitor.html", DefaultPlaces)
}

func GetDataMonitor(w http.ResponseWriter, r *http.Request) {
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

	c := session.DB(database).C(collection)

	dbSize, err := c.Count()

	if err != nil {
		log.Fatal(err)
	}

	var sensor_data Sensor

	err = c.Find(nil).Skip(dbSize - 1).One(&sensor_data)

	if err != nil {
		log.Fatal(err)
	}

	f64from32 := float64(sensor_data.Distance)

	hasil_conv := fmt.Sprintf("%s", FloatToString(f64from32))

	fmt.Fprintf(w, "{\"vrms\": "+hasil_conv+"}")
	//fmt.Println(hasil_conv)
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

	//fmt.Printf("TOPIC: %s\n", msg.Topic())
	//fmt.Printf("MSG: %s\n", msg.Payload())
	//fmt.Println(time.Now())

	sensing := fmt.Sprintf("%s", msg.Payload())
	res := Sensor{}
	json.Unmarshal([]byte(sensing), &res)

	res.Date = time.Now()

	c := session.DB(database).C(collection)
	err = c.Insert(res)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	fs2 := http.FileServer(http.Dir("js"))
	http.Handle("/js/", http.StripPrefix("/js/", fs2))
	http.HandleFunc("/", HomeTemplate)
	http.HandleFunc("/insert", InsertTemplate)
	http.HandleFunc("/monitor", MonitoringTemplate)
	http.HandleFunc("/getdata", GetDataMonitor)

	runtime.GOMAXPROCS(1)

	var wg sync.WaitGroup
	wg.Add(3)

	opts := MQTT.NewClientOptions().AddBroker(hostMQTT)
	opts.SetClientID("jarinsos")
	opts.SetDefaultPublishHandler(f)

	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	go func() {
		defer wg.Done()

		log.Println("Mqtt start at port 1883")
		if token := c.Subscribe("flood", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
	}()

	go func() {
		defer wg.Done()

		log.Println("Web Service start at port 8000")
		http.ListenAndServe(webappPort, nil)
	}()

	wg.Wait()
}
