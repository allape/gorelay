package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

var HostAddr = "0.0.0.0:8080"

// var PowerPin = "78"
// var LightPin = "76"

var OperatedPins = []string{}

func CleanUpGPIO() {
	for _, pin := range OperatedPins {
		err := exec.Command(
			fmt.Sprintf("echo %s > /sys/class/gpio/unexport", pin),
		).Run()
		if err != nil {
			log.Println("Failed to clean up pin", pin, ", because (of)", err)
		} else {
			log.Println("Clean up GPIO", pin)
		}
	}
}

func SetupGPIO() {
	state, err := host.Init()
	if err != nil {
		panic(err)
	}
	log.Printf("\nCurrent State: %v \n", state)

	// all := gpioreg.All()
	// for index, pin := range all {
	// 	state := gpio.Low
	// 	for i := 0; i < 4; i++ {
	// 		fmt.Printf("Set #%v/%v to %v\n", index, pin, state)
	// 		err := pin.Out(state)
	// 		if err != nil {
	// 			fmt.Printf("Unable to set %v to %v because (of) %v\n", pin, state, err)
	// 			break
	// 		}
	// 		time.Sleep(500 * time.Millisecond)
	// 		state = !state
	// 	}
	// }

	// p := gpioreg.ByName("78")
	// t := time.NewTicker(3 * time.Second)
	// for l := gpio.Low; ; l = !l {
	// 	err := p.Out(l)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("Set %v to %v\n", p, l)
	// 	<-t.C
	// }
}

func SetupHTTPServer() {
	router := gin.Default()
	router.Use(cors.Default())

	router.PUT("/pin/:pin/:status", func(ctx *gin.Context) {
		pinNumber := ctx.Params.ByName("pin")
		statusStr := ctx.Params.ByName("status")
		status := gpio.Low
		if statusStr == "1" {
			status = gpio.High
		}
		pin := gpioreg.ByName(pinNumber)
		if pin == nil {
			ctx.JSON(http.StatusNotFound, false)
			return
		}
		err := pin.Out(status)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}
		if !slices.Contains(OperatedPins, pinNumber) {
			OperatedPins = append(OperatedPins, pinNumber)
		}
		ctx.JSON(http.StatusOK, true)
	})

	router.Run(HostAddr)
}

func SetupCtrlC() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	log.Println("\nReceived signal:", sig)
	CleanUpGPIO()
	os.Exit(0)
}

func main() {
	SetupGPIO()
	go SetupHTTPServer()
	SetupCtrlC()
}
