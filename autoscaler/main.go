package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {

	periodStr, ok := os.LookupEnv("AUTOSCALER_CHECK_PERIOD")

	if !ok {
		error := errors.New("autoscaler check period environment varible not set")
		log.Fatal(error)
	}

	period, _ := strconv.Atoi(periodStr)

	for t := range time.NewTicker(time.Duration(period) * time.Second).C {
		fmt.Printf("Hello from Autoscaler at time %s", t)
	}
}
