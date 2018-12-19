package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

var version string

type GameData struct {
	Username string     `json:"username"`
	Status   StatusData `json:"status"`
}

type StatusData struct {
	Status []bool `json:"status"`
}

// TODO: refactor these global variables to use a game class to store game status and parameters

// globalCurrentSensorValues contains the latest sensor values
// it is set by the sensor function, and read by the GET /status call
var globalCurrentSensorValues []int

// gameInProgress stores the game status (if it is in progress or not)
var gameInProgress = false

func main() {
	// Init logs
	logLevelRaw := os.Getenv("BEERPONG_LOG_LEVEL")
	if logLevelRaw == "" {
		logLevelRaw = "info"
	}

	logLevel, errParseLevel := log.ParseLevel(logLevelRaw)
	if errParseLevel != nil {
		log.WithError(errParseLevel).Fatal("error parsing log level")
	}
	log.SetLevel(logLevel)

	// Print version info
	log.WithField("version", version).Info("program version")

	// Get config raw
	usbPort := os.Getenv("BEERPONG_USB_PORT")
	baudRateRaw := os.Getenv("BEERPONG_BAUD_RATE")
	publicHtml := os.Getenv("BEERPONG_PUBLIC_HTML")
	mockRaw := os.Getenv("BEERPONG_MOCK")
	productionRaw := os.Getenv("BEERPONG_PRODUCTION")
	listenPort := os.Getenv("BEERPONG_LISTEN_PORT")

	// Get default values
	if usbPort == "" {
		usbPort = "/dev/ttyACM0"
	}
	if baudRateRaw == "" {
		baudRateRaw = "115200"
	}
	if publicHtml == "" {
		publicHtml = "../rpi-front/dist"
	}
	if mockRaw == "" {
		mockRaw = "false"
	}
	if productionRaw == "" {
		productionRaw = "false"
	}
	if listenPort == "" {
		listenPort = "8080"
	}

	// Parse raw values
	baudRate, errAtoi := strconv.Atoi(baudRateRaw)
	if errAtoi != nil {
		log.WithError(errAtoi).Fatal("error parsing baud rate value")
	}
	mock, errParseBool := strconv.ParseBool(mockRaw)
	if errParseBool != nil {
		log.WithError(errParseBool).Fatal("error parsing mock value")
	}
	production, errParseBool := strconv.ParseBool(productionRaw)
	if errParseBool != nil {
		log.WithError(errParseBool).Fatal("error parsing production value")
	}

	// Print configuration info
	log.WithFields(log.Fields{
		"logLevel":   logLevelRaw,
		"usbPort":    usbPort,
		"baudRate":   baudRate,
		"publicHtml": publicHtml,
		"mock":       mock,
		"production": production,
		"listenPort": listenPort,
	}).Info("current configuration")

	// Configure Gin release mode
	if production {
		gin.SetMode(gin.ReleaseMode)
	}

	// Test database connection
	getDbClient()
	// Close on exit
	defer getDbClient().close()

	// Init buffer data variables
	buf := make([]byte, 128)
	var stringList []string

	// Init Web server/API
	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile(publicHtml, true)))

	router.GET("/version", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"version": version})
	})

	router.GET("/status", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": globalCurrentSensorValues})
	})

	router.POST("/begin", func(ctx *gin.Context) {
		var gameDataJson GameData
		if err := ctx.ShouldBindJSON(&gameDataJson); err != nil {
			log.WithError(err).Error("error on JSON binding")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userExists, errUserExists := getDbClient().userExists(gameDataJson.Username)
		if errUserExists != nil {
			log.WithError(errUserExists).Error("error on user existence")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": errUserExists.Error()})
			return
		}

		if !userExists {
			username, score, errInsertUser := getDbClient().insertUser(gameDataJson.Username, 0)
			if errInsertUser != nil {
				log.WithError(errInsertUser).Error("error on user insertion")
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": errInsertUser.Error()})
				return
			}
			log.WithFields(log.Fields{
				"username": username,
				"score":    score,
			}).Info("user insertion")
		}

		gameInProgress = true
		ctx.JSON(http.StatusOK, gin.H{"status": "begin"})
	})

	router.POST("/end", func(ctx *gin.Context) {
		var gameDataJson GameData
		if err := ctx.ShouldBindJSON(&gameDataJson); err != nil {
			log.WithError(err).Error("error on JSON binding")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !gameInProgress {
			log.WithField("gameInProgress", gameInProgress).Error("the game is not started")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "the game is not started"})
			return
		}

		userExists, errUserExists := getDbClient().userExists(gameDataJson.Username)
		if errUserExists != nil {
			log.WithError(errUserExists).Error("error on user existence")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": errUserExists.Error()})
			return
		}

		if !userExists {
			log.WithField("gameDataJson.Username", gameDataJson.Username).Error("this user does not exist")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "this user does not exist"})
			return
		}

		score := 0
		for _, ballPresent := range gameDataJson.Status.Status {
			if ballPresent {
				score++
			}
		}

		username, score, errUpdateUser := getDbClient().updateUser(gameDataJson.Username, score)
		if errUpdateUser != nil {
			log.WithError(errUpdateUser).Error("error on user update")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error on user update"})
			return
		}
		log.WithFields(log.Fields{
			"username": username,
			"score":    score,
		}).Info("user update")

		gameInProgress = false
		ctx.JSON(http.StatusOK, gin.H{"status": "end", "result": globalCurrentSensorValues})
	})

	httpServer := &http.Server{
		Addr:           ":" + listenPort,
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if !mock {
		// USB connection
		c := &serial.Config{Name: usbPort, Baud: baudRate}
		s, errOpenPort := serial.OpenPort(c)
		if errOpenPort != nil {
			log.WithError(errOpenPort).Fatal("error when opening USB connection")
		}

		// Infinite loop to read USB, parse and handle the data
		go func() {
			for {
				n, errRead := s.Read(buf)
				if errRead != nil {
					// In case of error, just print the error and continue reading without handling any data
					log.WithError(errRead).Error("error when reading on USB")
					continue
				}
				log.WithField("buf[:n]", fmt.Sprintf("%s", buf[:n])).Debug("buffer read")

				stringList = append(stringList, parseBuffer(buf[:n], []string{})...)
				log.WithField("stringList", stringList).Debug("buffer parsed")

				stringList = handleStringList(stringList)
				log.WithField("stringList", stringList).Debug("buffer strings handled")
			}
		}()
	} else {
		// Handle mock mode
		dbTest()
		handleMocks(buf, stringList)
	}

	// Start HTTP server
	log.WithField("address", httpServer.Addr).Info("HTTP server running")

	if err := httpServer.ListenAndServe(); err != nil {
		log.WithError(err).Error("error while running ListenAndServe")
	}
}

// parseBuffer parses the buffer into an array of strings using the '\r' character as a separator and stripped of '\n' characters
func parseBuffer(buffer []byte, stringAcc []string) []string {
	var relevantBytes []byte

	for index, byteValue := range buffer {
		if byteValue == 13 {
			// If we encounter a '\r' character, divide the buffer slice in two
			// and do a recursive call to finally get a parsed string array
			stringAcc = parseBuffer(buffer[:index], stringAcc)
			stringAcc = parseBuffer(buffer[index+1:], stringAcc)
			return stringAcc
		} else if byteValue != 10 {
			// If we encounter a '\n' character, do not parse it into a string, just ignore it
			relevantBytes = append(relevantBytes, byteValue)
		}
	}

	// If an empty string is found in the end, just return the accumulator as is
	if string(relevantBytes) == "" {
		return stringAcc
	}

	// Add the string found into the accumulator
	return append(stringAcc, string(relevantBytes))
}

// handleStringList computes the parsed buffer until it finds a pattern matching string to eliminate it from the parsed buffer array
func handleStringList(stringList []string) []string {
	var stringExpression string
	var lastIndexMatch int
	var patternFound bool
	funcRegexp := regexp.MustCompile("^.*?\\(.*?\\);$")

	// Go through each string in the parsed buffer array
	for index, stringValue := range stringList {
		// Append the latest value to the general string
		stringExpression += stringValue

		if funcRegexp.MatchString(stringExpression) {
			// If the computed string matches the pattern, we raise the patternFound flag
			// We also mark the index where the string was completed and handled to eliminate it after
			patternFound = true
			lastIndexMatch = index

			// Handle the function string
			handleFunctionExpression(stringExpression)

			// Reset the general string to continue to compute the buffer array
			stringExpression = ""
		}
	}

	// If nothing has been found, return the parsed buffer as is
	if !patternFound {
		return stringList
	}

	// If some functions has been found, return the parsed buffer stripped of the processed strings
	// so that we just leave the unprocessed ones for further processing
	return stringList[lastIndexMatch+1:]
}

func handleFunctionExpression(functionString string) {
	log.WithField("functionString", functionString).Debug("function match")

	funcRegexp := regexp.MustCompile("^(.*?)\\((.*?)\\);$")
	functionComposition := funcRegexp.FindStringSubmatch(functionString)

	// Get function information
	functionName := functionComposition[1]
	functionArgs := strings.Split(functionComposition[2], ",")

	// Act differently following the function name
	switch functionName {
	case "sensor":
		log.Debug("sensor function")
		var sensorArgs []int

		for _, arg := range functionArgs {
			argInt, errAtoi := strconv.Atoi(arg)
			if errAtoi != nil {
				// In case of error, just print the error and stop the function execution without handling any data
				log.WithError(errAtoi).Error("unexpected error during sensor parameters conversion")
				return
			}
			sensorArgs = append(sensorArgs, argInt)
		}

		// Handle the sensors
		sensor(sensorArgs)
	default:
		log.WithField("functionString", functionString).Warn("no corresponding function")
	}
}

func sensor(values []int) {
	log.WithField("values", values).Info("update current sensor values")
	globalCurrentSensorValues = values
}
