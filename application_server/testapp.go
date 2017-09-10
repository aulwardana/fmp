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

	_ "github.com/lib/pq" //external library for postgresql
	"gopkg.in/mgo.v2"     //external library for mongodb

	MQTT "github.com/eclipse/paho.mqtt.golang" //external library for interfacing with Mosquito MQTT
)

const (

	//PostgreSQL database configuration
	hosts      = "your_localhost:port"
	database   = "your_db_name"
	username   = "your_db_username"
	password   = "your_db_password"
	collection = "your_db_collection"

	//MongoDB database configuration
	host    = "your_localhost"
	port    = psql_port
	user    = "your_db_username"
	pwd     = "your_db_password"
	dbname  = "your_db_name"
	sslmode = "disable" // default ssl value

	//Default MQTT URL configuration
	hostMQTT = "tcp://localhost:1883"

	webappPort = ":8000" // default web app port

	//Set default location for taging sensor
	DefLatitude  = "your_latitude"
	DefLongitude = "your_longitude"
)

//Define template for every page that user can access
var templateHome = template.Must(template.ParseFiles("index.html"))
var templateInsert = template.Must(template.ParseFiles("insert.html"))
var templateMonitoring = template.Must(template.ParseFiles("monitor.html"))

//Json sensor structure for insert data from mqtt payload message to mongodb
type Sensor struct {
	Code     string    `json:"Code"`
	Distance float32   `json:"Distance"`
	Date     time.Time `json:"Date"`
}

//Tagging sensor structure data
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

//Default sensor structure for initializing maps value
type DefaultPlace struct {
	Latitude  string
	Longitude string
}

//Function for convert float to string format
func FloatToString(input_num float64) string {
	return strconv.FormatFloat(input_num, 'f', 2, 64)
}

//Function for handle and control activity in home page
func HomeTemplate(w http.ResponseWriter, r *http.Request) {

	//PostgreSQL info configuration for connect with posgresql database
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, pwd, dbname, sslmode)

	//Connect to postgresql database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//Test connection with ping
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	Places := []Place{} //array for save multiple data from querying

	//Select query table in postgresql database
	rows, err := db.Query("SELECT * FROM public.place")
	if err != nil {
		http.Error(w, "Select error", 500)
		return
	}

	//Scan for querying every row
	for rows.Next() {
		c := Place{}
		err := rows.Scan(&c.Uid, &c.Name, &c.Type, &c.Location, &c.State, &c.Kawasan, &c.Latitude, &c.Longitude, &c.Code)
		if err != nil {
			http.Error(w, "Decompose error", 500)
			return
		}
		//Append multiple data from querying
		Places = append(Places, c)
	}

	//Append Palces data to home page maps
	templateHome.ExecuteTemplate(w, "index.html", Places)
}

//Function for handle and control activity in insert page
func InsertTemplate(w http.ResponseWriter, r *http.Request) {

	//If insert page access with "GET" method, then show the insert page with default maps value
	if r.Method != "POST" {
		DefaultPlaces := DefaultPlace{DefLatitude, DefLongitude}
		templateInsert.ExecuteTemplate(w, "insert.html", DefaultPlaces)
		return
	}

	//Get value from "POST" method that send from insert page
	nameJ := r.FormValue("name")
	typeJ := r.FormValue("type")
	locationJ := r.FormValue("location")
	stateJ := r.FormValue("state")
	kawasanJ := r.FormValue("kawasan")
	latitudeJ := r.FormValue("lat")
	longitudeJ := r.FormValue("long")
	codeJ := r.FormValue("code")

	//PostgreSQL info configuration for connect with posgresql database
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, pwd, dbname, sslmode)

	//Connect to postgresql database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//Test connection with ping
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	//Insert data from insert page to postgresql database
	var lastInsertId int
	err = db.QueryRow("INSERT INTO public.place(name, type, location, state, kawasan, latitude, longitude, code) VALUES($1,$2,$3,$4,$5,$6,$7,$8) returning uid;", nameJ, typeJ, locationJ, stateJ, kawasanJ, latitudeJ, longitudeJ, codeJ).Scan(&lastInsertId)
	log.Println("last inserted id =", lastInsertId)

	//Handling error and success insert
	if err != nil {
		http.Error(w, "Insert error", 500)
		return
	} else {
		http.Redirect(w, r, "/", 301)
		return
	}
}

//Function for handle and control activity in monitor page
func MonitoringTemplate(w http.ResponseWriter, r *http.Request) {
	DefaultPlaces := DefaultPlace{DefLatitude, DefLongitude}
	templateMonitoring.ExecuteTemplate(w, "monitor.html", DefaultPlaces)
}

func GetDataMonitor(w http.ResponseWriter, r *http.Request) {

	//MongoDB info configuration for connect with mongodb database
	info := &mgo.DialInfo{
		Addrs:    []string{hosts},
		Timeout:  60 * time.Second,
		Database: database,
		Username: username,
		Password: password,
	}

	//Start connection with mongodb
	session, err := mgo.DialWithInfo(info)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB(database).C(collection) //open database and collection

	dbSize, err := c.Count() //check the db size

	//Handling db size nil
	if err != nil {
		log.Fatal(err)
	}

	var sensor_data Sensor //create variable with sensor structure

	err = c.Find(nil).Skip(dbSize - 1).One(&sensor_data) //get the latest monitoring value data

	//Handling nil value from querying
	if err != nil {
		log.Fatal(err)
	}

	f64from32 := float64(sensor_data.Distance) //convert from float 32 to 64

	hasil_conv := fmt.Sprintf("%s", FloatToString(f64from32)) //convert from float to string format

	fmt.Fprintf(w, "{\"vrms\": "+hasil_conv+"}") //print data with JSON format
}

var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {

	//MongoDB info configuration for connect with mongodb database
	info := &mgo.DialInfo{
		Addrs:    []string{hosts},
		Timeout:  60 * time.Second,
		Database: database,
		Username: username,
		Password: password,
	}

	//Start connection with mongodb
	session, err := mgo.DialWithInfo(info)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	//Capture MQTT message payload from mqtt broker
	sensing := fmt.Sprintf("%s", msg.Payload())

	/*
		Unmarshal JSON format from sensing data in mqtt payload message.
		Convert the message to Json sensor structure for insert data to mongodb.
	*/
	res := Sensor{}
	json.Unmarshal([]byte(sensing), &res)
	res.Date = time.Now() //adding date and time from server to data

	//Insert reformating data from mqtt payload message to database
	c := session.DB(database).C(collection)
	err = c.Insert(res)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	/*
		Add default asset like javascript and CSS library for frontend.
		The defult folder will be serve with web application in first start.
	*/
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	fs2 := http.FileServer(http.Dir("js"))
	http.Handle("/js/", http.StripPrefix("/js/", fs2))

	//Define all monitoring page
	http.HandleFunc("/", HomeTemplate)
	http.HandleFunc("/insert", InsertTemplate)
	http.HandleFunc("/monitor", MonitoringTemplate)

	//Supply sensor data from database to realtime chart in monitor page
	http.HandleFunc("/getdata", GetDataMonitor)

	//Manage multiple process for mqtt subscriber and http web app
	runtime.GOMAXPROCS(1)
	var wg sync.WaitGroup
	wg.Add(3)

	//Mqtt preconfiguration
	opts := MQTT.NewClientOptions().AddBroker(hostMQTT)
	opts.SetClientID("jarinsos") //the default client id is "jarinsos", you can change with your own client id
	opts.SetDefaultPublishHandler(f)

	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	//Serve and control mqtt subscriber
	go func() {
		defer wg.Done()
		log.Println("Mqtt start at port 1883")
		//the default topic is "flood", you can change with your own topic.
		if token := c.Subscribe("flood", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
	}()

	//Serve and control http web app
	go func() {
		defer wg.Done()
		log.Println("Web Service start at port 8000")
		http.ListenAndServe(webappPort, nil)
	}()

	wg.Wait()
}
