package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

var version string

// globalCurrentSensorValues contains the latest sensor values
// it is set by the sensor function, and read by the GET /status call
var globalCurrentSensorValues []int

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

	// Parse raw values
	baudRate, errAtoi := strconv.Atoi(baudRateRaw)
	if errAtoi != nil {
		log.WithError(errAtoi).Fatal("error parsing baud rate value")
	}
	mock, errParseBool := strconv.ParseBool(mockRaw)
	if errParseBool != nil {
		log.WithError(errParseBool).Fatal("error parsing mock value")
	}

	// Print configuration info
	log.WithFields(log.Fields{
		"logLevel":   logLevelRaw,
		"usbPort":    usbPort,
		"baudRate":   baudRate,
		"publicHtml": publicHtml,
		"mock":       mock,
	}).Info("current configuration")

	// Test database connection
	getDbClient()
	// Close on exit
	defer getDbClient().close()

	// Init buffer data variables
	buf := make([]byte, 128)
	var stringList []string

	// Handle mock mode
	if mock {
		dbTest()
		handleMocks(buf, stringList)
		return
	}

	// USB connection
	c := &serial.Config{Name: usbPort, Baud: baudRate}
	s, errOpenPort := serial.OpenPort(c)
	if errOpenPort != nil {
		log.WithError(errOpenPort).Fatal("error when opening USB connection")
	}

	// Infinite loop to read USB, parse and handle the data
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
	functionArgs := strings.Split(functionComposition[2], ", ")

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
