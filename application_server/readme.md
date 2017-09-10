## Platform Architecture
!["Image 1"](https://s19.postimg.org/7murefcsz/Software_Architecture_of_Server.jpg)


## Application Preview
You can see FPM (Flood Monitoring Platform) application preview in [wiki FPM](# "wiki FPM").


## Server Requirement
1. VM with 1 CPU, 1 GB Ram, 10 GB HDD
2. Redhat Linux OS (centos, fedora, or redhat enterprise)
3. SSH Remote Active (for deploy)


## Prerequisite Environment
Before you deploy this platform to cloud, you must install this application :
1. **MongoDB**, you can open port 27017 in public to remote (recomended use robomongo). Please add username and password in collection that you use for this platform.
2. **PostgreSQL**, you can open port 5432 in public to remote (recomended use pgadmin). Please add username and password in table that you use for this platform.
3. **Mosquito MQTT**, this application running in port 1883. The hardware will send sensing data through this port.
3. **Nginx**, load the platform in `localhost:8000` for default url and port.


## Golang External Library
1. [lib/pq](https://github.com/lib/pq "lib/pq")
2. [mgo.v2](http://gopkg.in/mgo.v2 "mgo.v2")
3. [paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang "paho.mqtt.golang")