package UsageGenerator

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/dowts/go-packages/AWSLambda"
	"github.com/dowts/go-packages/AzureFunctions"
	"github.com/dowts/go-packages/GoogleFunctions"
	"github.com/dowts/go-packages/IBMFunctions"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/rand"
)

// User configured parameters
var botnetSize int

// Array of botnet IP addresses
var botnetIPs []string

// attackers
var botBursty = false

// attacks
// 1 = constant
// 2 = exponential
// 3 = random
var attackchoice = 1

// ip changing
// 0 = A
// 1 = B
// 2 = C
// 3 = D  NOT APPLICABLE
// 4 = BC
// 5 = BD
// 6 = CD
// 7 = BCD
var attackIPchoice = 0

var attacknum = 0

// D
// how many bots from available pool to use
var botnetSlice int
var botnetPoolIPs []string

// 1
// Constant rate attack
// 200 requests per hour times 24 as 24 hours in one day (timestep set to 3)
var constantAttacknum = 2000

// 2
// Exponential rate attack
// start requests at 10 and increase by factor of 1.005 per timestep
var expoAttacknum = 10.0
var expoAttackfactor = 1.05

// 3
// Random rate attack
// random request rate between 0 and 2000 per hour
var randAttacknum = 0
var randAttackFactor = 2000

// variable to keep track of bot request so IPs can be changed
var numRequests = 0

// reset counter to cancel attacks in progress that are going too long after request limit reached
var resetCount = 0

// reset value to reset IPs after amount of time steps
var resetVal = 1

// Bot traffic generator
func botTraffic(waitG *sync.WaitGroup, logs *[]interface{}, csv *string, startTime time.Time, attack int) {
	defer waitG.Done()

	if len(botnetIPs) < botnetSlice {
		time.Sleep(5 * time.Second)
	}
	ipaddress := botnetIPs[rand.Intn(botnetSlice)]

	dif := rand.Intn(timeFactor)
	timestamp := startTime.Add(time.Millisecond * time.Duration(dif))

	//B
	//keep track of reset count when attack started
	startCount := resetCount

	switch attack {
	//1
	//constant rate attack
	case 1:
		attacknum = constantAttacknum
	//2
	//Exponential attack rate
	case 2:
		attacknum = int(expoAttacknum)
	//3
	//Random attack rate
	case 3:
		attacknum = randAttacknum

	default:

	}

	if botBursty {
		if rand.Intn(100) > 25 {
			attacknum = attacknum / 10
		}
	}
	for a := 0; a < attacknum; a++ {
		index := rand.Intn(len(functionChains))
		//B
		if attackIPchoice == 1 || attackIPchoice == 4 || attackIPchoice == 5 || attackIPchoice == 7 {
			//reset IP address after request limit reached
			if resetCount > startCount {
				if len(botnetIPs) < botnetSlice {
					time.Sleep(5 * time.Second)
				}
				ipaddress = botnetIPs[rand.Intn(botnetSlice)]
				startCount += 1
			}
		}
		for i, value := range functionChains[index] {
			if i != 0 {

				dif := functions[value].Runtime + uint64(rand.Intn(1000))
				timestamp = timestamp.Add(time.Millisecond * time.Duration(dif))
				AWSLambda.RunFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)
				GoogleFunctions.RunFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)
				AzureFunctions.RunFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)
				IBMFunctions.RunFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)

				//B
				if attackIPchoice == 1 || attackIPchoice == 4 {
					mu.Lock()
					//increment number of requests
					numRequests += 1

					//if number of requests reached by botnet is over 5000, reset IPs
					if numRequests >= 5000 {
						numRequests = 0
						botnetIPs = botnetIPs[:0]
						//generated IP addresses for the botnet
						for ip := 0; ip < botnetSlice; ip++ {
							botnetIPs = append(botnetIPs, genIpaddr())

						}
						resetCount += 1
						//ipaddress = botnetIPs[rand.Intn(botnetSize)]

					}
					mu.Unlock()
				}

				//BD & BCD
				if attackIPchoice == 5 || attackIPchoice == 7 {
					mu.Lock()
					//increment number of requests
					numRequests += 1

					//if number of requests reached by botnet is over 5000, reset IPs
					if numRequests >= 5000 {
						numRequests = 0
						seed := rand.Intn(botnetSize - botnetSlice)
						botnetIPs = botnetPoolIPs[seed : seed+botnetSlice]
						resetCount += 1
						ipaddress = botnetIPs[rand.Intn(botnetSlice)]

					}
					mu.Unlock()
				}
				mu.Lock()
				if fileWrite == 1 {
					*csv += fmt.Sprintf("%s,true,%d,%s\n", ipaddress, functions[value].ID, timestamp.Format(time.RFC3339))
					if len(*csv) >= 10000 {
						dump := *csv
						*csv = ""
						mu.Unlock()

						f, _ := os.OpenFile("simdata.csv", os.O_APPEND, 0644)
						f.Write([]byte(dump))
						f.Close()
					} else {
						mu.Unlock()
					}

				} else {
					*logs = append(*logs, bson.D{{Key: "IP", Value: ipaddress}, {Key: "functioID", Value: functions[value].ID}, {Key: "timestamp", Value: timestamp}, {Key: "bot", Value: true}})
					//Check if log list is over size then begin to persist to database
					if len(*logs) >= 10000 {
						dump := *logs
						*logs = []interface{}{}
						mu.Unlock()
						_, insertErr := collection.InsertMany(ctx, dump)
						if insertErr != nil {
							log.Fatal(insertErr)
						}
					} else {
						mu.Unlock()
					}
				}
			}
		}
	}

}
