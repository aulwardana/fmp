//NodeMCU ESP8266 board library
#include <ESP8266WiFi.h>

//MQTT library
#include <PubSubClient.h>

/*
  This pin is used for HC-SR05 Ultrasonic sensor.
  The sensor will get water level value from dainage system,
  by measure the distance between water level and sensor.
*/
#define TRIGGER_PIN   5  //Port D1 in NodeMCU board
#define ECHO_PIN      4  //Port D2 in NodeMCU board

/*
  This pin is used for actuator, buzzer and LED.
  The actuator will active when the distance treshold is achieve.
  It means the water level is overflow and flood disaster is coming.
*/
#define BUZZER        15 //Port D8 in NodeMCU board
#define LED           14 //Port D7 in NodeMCU board

const char* ssid = "Hostspot Name"; //setup your ssid from wifi hotspot
const char* password = "Hostspot Password"; //setup your password from wifi hotspot
const char* mqtt_server = "MQTT Server Url"; //setup your cloud server url with installed Mosquito MQTT server

String codeSns = "1a"; //sensor code for tagging location in platform

//variable to construct mqtt payload message from sensing data 
long lastMsg = 0;
char msg[50];
int value = 0;

/*
  This device using WiFi for connect to cloud server via internet.
  The actuator will active when the distance treshold is achieve.
*/
WiFiClient espClient;
PubSubClient client(espClient);

void setup_wifi() {
  delay(100);
  Serial.print("Connecting to ");
  Serial.println(ssid);
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) 
  {
    delay(500);
    Serial.print(".");
  }
  randomSeed(micros());
  Serial.println("");
  Serial.println("WiFi connected");
  Serial.println("IP address: ");
  Serial.println(WiFi.localIP());
}

void reconnect() {
  while (!client.connected()) 
  {
    Serial.print("Attempting MQTT connection...");
    String clientId = "SNS1A";
    clientId += String(random(0xffff), HEX);
    //if you MQTT broker has clientID, username, and password
    //please change following line to "if (client.connect(clientId,userName,passWord))""
    if (client.connect(clientId.c_str()))
    {
      Serial.println("connected");
      client.subscribe("flood");
    } else {
      Serial.print("failed, rc=");
      Serial.print(client.state());
      Serial.println(" try again in 5 seconds");
      delay(6000);
    }
  }
}

void setup() {
  Serial.begin (115200);
  setup_wifi();
  client.setServer(mqtt_server, 1883);
  pinMode(TRIGGER_PIN, OUTPUT);
  pinMode(ECHO_PIN, INPUT);
  pinMode(BUZZER, OUTPUT);
  pinMode(LED, OUTPUT);

  digitalWrite(LED, LOW);
}

void loop() {
  
  /*
    Ultrasonic sensor will active sensing and check the water level distance.
    All data is save in variable before store to cloud server.
  */
  double duration, distance;
  
  digitalWrite(TRIGGER_PIN, LOW);
  delayMicroseconds(2);
  digitalWrite(TRIGGER_PIN, HIGH);
  delayMicroseconds(10); 
  digitalWrite(TRIGGER_PIN, LOW);
  
  duration = pulseIn(ECHO_PIN, HIGH);
  distance = (duration/2) / 29.1;

  //check water level distance, when treshold value achieve, actuator will be active
  if (distance <= 12) {
     tone(BUZZER, 1000); //buzzer active
     digitalWrite(LED, HIGH); //led active
   } else {
     noTone(BUZZER); //buzzer deactive
     digitalWrite(LED, LOW); //led deactive
   }
  
  if (!client.connected()) {
    reconnect();
  }

  /*
    The mqtt payload message will be contruct as JSON type.
    The data that send to the server is sensor code and water level distance.
  */
  client.loop();
  String msg="{\"Code\": \"";
  msg= msg+ codeSns;
  msg = msg+"\", \"Distance\": ";
  msg=msg+ distance ;
  msg=msg+"}";
  char message[65];
  msg.toCharArray(message,65);
  Serial.println(message);
  delay(2000);
  //default topic is "flood", you can change with your own topic
  client.publish("flood", message);
}
