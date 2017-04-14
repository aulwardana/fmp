#include <ESP8266WiFi.h>
#include <PubSubClient.h>

#define TRIGGER_PIN   5  //D1
#define ECHO_PIN      4  //D2 
#define BUZZER        15 //D8
#define LED           14 //D7

const char* ssid = "Hostspot Name";
const char* password = "Hostspot Password";
const char* mqtt_server = "alamat_mqtt_server";

WiFiClient espClient;
PubSubClient client(espClient);
long lastMsg = 0;
char msg[50];
int value = 0;

String codeSns = "1a";

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
    //if you MQTT broker has clientID,username and password
    //please change following line to    if (client.connect(clientId,userName,passWord))
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
} //end reconnect()

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
  double duration, distance;
  
  digitalWrite(TRIGGER_PIN, LOW);
  delayMicroseconds(2);
  digitalWrite(TRIGGER_PIN, HIGH);
  delayMicroseconds(10); 
  digitalWrite(TRIGGER_PIN, LOW);
  
  duration = pulseIn(ECHO_PIN, HIGH);
  distance = (duration/2) / 29.1;

  if (distance <= 12) {
     tone(BUZZER, 1000);
     digitalWrite(LED, HIGH);
   } else {
     noTone(BUZZER);
     digitalWrite(LED, LOW);
   }
  
  if (!client.connected()) {
    reconnect();
  }
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
  client.publish("flood", message);
}
