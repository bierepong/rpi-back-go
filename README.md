# rpi-back-go
Raspberry Pi Golang backend

## Routes

* GET / client
* GET /status Gets an array of sensor status [0, 0, 0, 0, 0, 0]
* POST /begin {"username": "", "email": "", "balls": 6} Begins a game => {"status": "begin"}
* POST /end End the current game => {"status": "end", "result": [0, 0, 0, 0, 0, 0]}


## Install:

 * go
 * gcc-arm-linux-gnueabihf