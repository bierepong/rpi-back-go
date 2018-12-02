package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

var mocks = []string{
	"sensor(15, 25, 55, ",
	"85, 95, 45);\r\n",
	"sensor(15, 25, 55, 85, 95, 45);\r\n",
	"sensor(15, 25, 55, 85, 95, 45",
	");\r\n",
	"sensor(15, 25, 55, 85, 95, 45);\r\n",
	"senso",
	"r(15, 25, 55, 85, 95, 45);\r\n",
	"sensor(15, 25, 55, 85, 95, 45);\r\n",
	"sensor(1",
	"5, 25, 55, 85, 95, 4);\r\n",
	"sensor(15, 25, 55, 85, 95, 4);\r\n",
	"sensor(15, 25, 55, 85, 9, 4);\r\n",
	"sensor(15, 25, 5, 85, 9, 4);\r\n",
	"sensor(15, 2, 5, 85, 9, 4);\r\n",
	"sensor(15, 2, 5, 85, 9, 4);",
	"\r\nsensor(15, 2, 5, 85, 9, 4);\r",
	"\nsensor(15, 2, 5, 85, 9, 4);\r\nsensor(15, 2",
	", 5, 85, 9, 4);\r\nsensor(15, 2, 5, 85, 9, 4);\r\nsensor(15, 2, 5, 85, 9, 4);\r\n",
}

var expectedValues = [][]int{
	{15, 25, 55, 85, 95, 45},
	{15, 25, 55, 85, 95, 45},
	{15, 25, 55, 85, 95, 45},
	{15, 25, 55, 85, 95, 45},
	{15, 25, 55, 85, 95, 45},
	{15, 25, 55, 85, 95, 45},
	{15, 25, 55, 85, 95, 4},
	{15, 25, 55, 85, 95, 4},
	{15, 25, 55, 85, 9, 4},
	{15, 25, 5, 85, 9, 4},
	{15, 2, 5, 85, 9, 4},
	{15, 2, 5, 85, 9, 4},
	{15, 2, 5, 85, 9, 4},
	{15, 2, 5, 85, 9, 4},
	{15, 2, 5, 85, 9, 4},
	{15, 2, 5, 85, 9, 4},
	{15, 2, 5, 85, 9, 4},
}

func handleMocks(buf []byte, stringList []string) {
	log.Warn("app is in mock mode")
	for _, stringMock := range mocks {
		buf = []byte(stringMock)
		log.WithField("buf", fmt.Sprintf("%s", buf)).Debug("buffer read")

		stringList = append(stringList, parseBuffer(buf, []string{})...)
		log.WithField("stringList", stringList).Debug("buffer parsed")

		stringList = handleStringList(stringList)
		log.WithField("stringList", stringList).Debug("buffer strings handled")
	}

	log.WithField("expectedValues", expectedValues).Info("expected mock values")
	log.WithField("globalCurrentSensorValues", globalCurrentSensorValues).Info("last mock value")
}