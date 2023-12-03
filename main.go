package main

import (
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"slices"
	"strings"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

var RootCmd = &cobra.Command{
	Use:   "gorelay [flags] pin=high pin=low",
	Short: "GPIO controller via HTTP with golang and periph.io",
	Args:  cobra.MinimumNArgs(0),
	Run:   start,
}

var HostAddr = "0.0.0.0:8080"

// var PowerPin = "78"
// var LightPin = "76"

var OperatedPins []string

func SetPin(pinNumber string, state gpio.Level) (gpio.PinIO, error) {
	pin := gpioreg.ByName(pinNumber)
	if pin == nil {
		return nil, nil
	}
	err := pin.Out(state)
	if err != nil {
		return pin, err
	}
	if !slices.Contains(OperatedPins, pinNumber) {
		OperatedPins = append(OperatedPins, pinNumber)
	}
	return pin, nil
}

func CleanUpGPIO() {
	for _, pin := range OperatedPins {
		err := os.WriteFile("/sys/class/gpio/unexport", []byte(pin+"\n"), 0o644)
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

	router.Static("/app", "/app")
	router.PUT("/pin/:pin/:state", func(ctx *gin.Context) {
		pinNumber := ctx.Params.ByName("pin")
		statusStr := ctx.Params.ByName("state")
		state := gpio.Low
		if statusStr == "1" {
			state = gpio.High
		}
		pin, err := SetPin(pinNumber, state)
		if pin == nil {
			ctx.JSON(http.StatusNotFound, false)
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, true)
	})

	err := router.Run(HostAddr)
	if err != nil {
		log.Fatalln("Unable to start server because (of)", err)
	}
}

func SetupCtrlC() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGABRT)
	sig := <-sigChan
	log.Println("\nReceived signal:", sig)
	CleanUpGPIO()
	log.Println("Goodbye~")
	os.Exit(0)
}

func SetupCommand() {
	RootCmd.PersistentFlags().StringVarP(&HostAddr, "addr", "a", HostAddr, "Listening address for HTTP server")
	err := RootCmd.Execute()
	if err != nil {
		log.Fatalln("Unable to parse arguments", err)
	}
}

func start(_ *cobra.Command, args []string) {
	SetupGPIO()
	go SetupHTTPServer()

	re, _ := regexp.Compile("^\\d+=(h(igh)?|l(ow)?|0|1)$")
	for _, pinAndDefaultOut := range args {
		pinAndDefaultOut = strings.ToLower(pinAndDefaultOut)
		if !re.MatchString(pinAndDefaultOut) {
			log.Println("Pattern", pinAndDefaultOut, "is not valid, skip...")
			continue
		}
		splitValues := strings.Split(pinAndDefaultOut, "=")
		pinNumber, stateString := splitValues[0], splitValues[1]
		state := gpio.Low
		if strings.HasPrefix(stateString, "h") || stateString[0] == '1' {
			state = gpio.High
		}
		pin, err := SetPin(pinNumber, state)
		if pin == nil {
			log.Println("No pin named", pinNumber, "found")
			continue
		}
		if err != nil {
			log.Println("Unable to set", pin, "to", state, ", because (of)", err)
			continue
		}
		log.Println("Set", pin, "to", state)
	}
}

func main() {
	go SetupCommand()
	SetupCtrlC()
}
