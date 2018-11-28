package main

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

var version string

func main() {
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
		"usbPort":    usbPort,
		"baudRate":   baudRate,
		"publicHtml": publicHtml,
		"mock":       mock,
	}).Info("current configuration")

	c := &serial.Config{Name: usbPort, Baud: baudRate}
	s, errOpenPort := serial.OpenPort(c)
	if errOpenPort != nil {
		log.WithError(errOpenPort).Fatal("error when opening USB connection")
	}

	buf := make([]byte, 128)
	n, errRead := s.Read(buf)
	if errRead != nil {
		log.WithError(errRead).Fatal("error when reading on USB")
	}

	log.Infof("%q", buf[:n])
}
