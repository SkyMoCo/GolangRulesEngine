package main

import (
	"time"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	// Rules Engine
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
	// RasberryPi GPIO
	"github.com/stianeikeland/go-rpio"
	// MQTT Client in the future
	// "github.com/eclipse/paho.mqtt.golang"
) 


type TrafficFacts struct {
	YellowLightStart time.Time
	YellowLightTimer int  // How long as yellow light been on in seconds.   If zero, light is not on
	GreenLightStart time.Time
	GreenLightTimer int   //
	RedLightStart time.Time
	RedLightTimer int    // How long as red light been on in seconds.   If zeo, light is not on
	Pedestrians bool   // Are there pedestrians waiting, once pressed this "latches" and starts the sequence
}

// Global Vars
var (
	// Use mcu pin 22, corresponds to GPIO 3 on the pi
	SwitchPin = rpio.Pin(3)
	GreenPin = rpio.Pin(11)
	trafficFacts = TrafficFacts {GreenLightTimer: 0,YellowLightTimer: 0, RedLightTimer:0,}
)


func (re TrafficFacts) SecondsSince(startTime time.Time) int {

	slog.Info(fmt.Sprintf("SecondsSince: Start: %s ", startTime))
	// Only works for times w/in the last 24 hours per some doc I read somewhere
	difference := startTime.Sub(time.Now())
	return int(difference.Seconds())

}

// Turn on Pin by giving map name
func (re TrafficFacts) TurnOnLightByName(pname string) string {
	//var pin int
	slog.Info(fmt.Sprintf("TurnOnLightByName: Name is: %s ", pname))

	// There has to be a better way than this.
	// Use switch on the pname variable. 
	switch { 
	case pname == "green":
		slog.Info("Green light on")
		thisPin := rpio.Pin(11)
		thisPin.High()
		trafficFacts.GreenLightStart = time.Now()
		trafficFacts.GreenLightTimer = 1
	case pname == "yellow":
		slog.Info("Yellow light on")
		thisPin := rpio.Pin(10)
		thisPin.High()
		trafficFacts.YellowLightStart = time.Now()
		trafficFacts.YellowLightTimer = 1

	case pname == "red":
		slog.Info("Red light on")
		thisPin := rpio.Pin(9)
		thisPin.High()
		trafficFacts.RedLightStart = time.Now()
		trafficFacts.RedLightTimer = 1

	}
	return fmt.Sprintf("%s light on", pname)
}

// Turn off Pin by giving map name
func (re TrafficFacts) TurnOffLightByName(pname string) string {
	//var pin int
	fmt.Println("piname is ", pname)
	slog.Info(fmt.Sprintf("TurnOffLightByName: Name is: %s ", pname))
	
	// This is testing
	// Map of pin names to values
	//	pins := map[string]int{
	//	"red": 9,
	//	"yellow": 10,
	//	"green": 11,
	//}
	// There has to be a better way than this.
	// Use switch on the pname variable. 
	switch { 
	case pname == "green":
		slog.Info("Green light off")
		thisPin := rpio.Pin(11)
		thisPin.Low()
		trafficFacts.GreenLightTimer = 0
	case pname == "yellow":
		slog.Info("Yellow light off")
		thisPin := rpio.Pin(10)
		thisPin.Low()
		trafficFacts.YellowLightTimer = 0
	case pname == "red":
		slog.Info("Red light off")
		thisPin := rpio.Pin(9)
		thisPin.Low()
		trafficFacts.RedLightTimer = 0
	}

	return fmt.Sprintf("%s light off", pname)
}


// Update the switch state every 250 milliseconds, a "poor" mans event handler
func readSwitch(switchPin rpio.Pin) {
	ticker := time.NewTicker(time.Millisecond * 250)

	for range ticker.C {
		if SwitchPin.Read() == 0 {
			trafficFacts.Pedestrians = true
			fmt.Println("Pedestrians are waiting")
		}
	}
}

// Run the rules engine every X seconds
func runRules(knowledgeBase *ast.KnowledgeBase) {

	tocker := time.NewTicker(time.Second * 2)   // Run X times per second
	for range tocker.C {
		slog.Info("Reset Facts")
		if trafficFacts.Pedestrians == true {
			fmt.Println("Pedestrians is true")
		} else {
			fmt.Println("Pedestrians is false")
		}
		if trafficFacts.GreenLightTimer > 0 {
			fmt.Println("GreenLight is on")
		} else {
			fmt.Println("GreenLight is off")
		}
		// Since we are simulating outside facts coming in, we can "fake" a timer running on the lights
		// This way we don't assume the rules will run every set period.

		if trafficFacts.GreenLightTimer > 0 {
			elapsedTime := time.Since(trafficFacts.GreenLightStart)
			trafficFacts.GreenLightTimer = int(elapsedTime.Seconds())
		}
		if trafficFacts.YellowLightTimer > 0 {
      elapsedTime := time.Since(trafficFacts.YellowLightStart)
      trafficFacts.YellowLightTimer = int(elapsedTime.Seconds())
			fmt.Println(trafficFacts.YellowLightTimer);
		}
		if trafficFacts.RedLightTimer > 0 {
      elapsedTime := time.Since(trafficFacts.RedLightStart)
      trafficFacts.RedLightTimer = int(elapsedTime.Seconds())

		}
		slog.Info("Run the rules engine")
		dataCtx := ast.NewDataContext()
		err := dataCtx.Add("TF", trafficFacts)
		if err != nil {
			panic(err)
		}
			
		engine := &engine.GruleEngine{MaxCycle: 10}
		err = engine.Execute(dataCtx, knowledgeBase)
		if err != nil {
			panic(err)
		}
	}
}



// Load the rules and return a Knowledge Base of them.
func loadRules() *ast.KnowledgeBase {
	
	knowledgeLibrary := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(knowledgeLibrary)

	// Now add the rules from a file, each file contains a group of rules
	fileRes := pkg.NewFileResource("rules.grl")
	err := ruleBuilder.BuildRuleFromResource("TrafficRules", "0.0.1", fileRes)
	if err != nil {
    panic(err)
	}

	knowledgeBase, err := knowledgeLibrary.NewKnowledgeBaseInstance("TrafficRules", "0.0.1")
	if err != nil {
		panic(err)
	}
	// Once all the rules are added into our Knowledge Base,  return it
	return knowledgeBase
}
	
func main() {
	// Create knowledgeBase variable
	var knowledgeBase *ast.KnowledgeBase
	
	// Set our log level above the default warn...
	slog.SetLogLoggerLevel(slog.LevelInfo)

	// Setup the gpio pins by creating a rpio structure
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rpio.Pin(9).Output()
	rpio.Pin(10).Output()
	rpio.Pin(11).Output()
	rpio.Pin(9).Low()
	rpio.Pin(10).Low()
	rpio.Pin(11).Low()

	
	go readSwitch(rpio.Pin(3))
	
	slog.Info("Load rules")
	knowledgeBase = loadRules()
	
	// These facts should come from outside, like a thread that reads a message queue...
	slog.Info("Update Facts")

	// Facts are then loaded into a data context, we could have a whole bunch of facts
	//	dataCtx := ast.NewDataContext()
	//err := dataCtx.Add("TF", trafficFacts)
	//if err != nil {
  //  panic(err)
	//}
	
	slog.Info("Kick off the rule engine")
	go runRules(knowledgeBase)

	// Block until forever
	done := make(chan os.Signal, 1)
  signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
  fmt.Println("Blocking, press ctrl+c to continue...")
  <-done  // Will block here until user hits ctrl+c
	
	rpio.Pin(9).Low()
	rpio.Pin(10).Low()
	rpio.Pin(11).Low()
	rpio.Close()
	// Clean up on ctrl-c and turn lights out
	os.Exit(0)

}


