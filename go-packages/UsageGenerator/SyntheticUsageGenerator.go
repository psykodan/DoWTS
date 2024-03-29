package UsageGenerator

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/dowts/go-packages/AWSLambda"
	"github.com/dowts/go-packages/AzureFunctions"
	"github.com/dowts/go-packages/GoogleFunctions"
	"github.com/dowts/go-packages/IBMFunctions"
	"github.com/dowts/go-packages/function"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

// User configured parameters
var numFunctions, userbaseSize int
var timestep, numSteps, usersPerStep, trafficPerStep int

// divisor based on time step
var timeFactor int

// Array of user IP addresses
var ipaddresses []string

// array of functions
var functions []function.Function

// graphs of functions that execute one after another
var functionChains [][]int

// wait group for synchronised goroutines
var wg sync.WaitGroup

// Mutex for writing logs
var mu sync.Mutex

// Attack parameters
// bursty traffic
// real
var realBursty = true

var attackStart int
var attackDur int

func SyntheticDataGenerator() {

	//TRAFFIC MODELLING

	//Poisson function random number
	srcArrive := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	poisson := distuv.Poisson{Lambda: float64(trafficPerStep), Src: srcArrive}

	//start time for logs, start at midnight
	//t := time.Now().Add(-time.Hour * 24).Format("2006/01/02")
	startTime, _ := time.Parse("2006/01/02", "2020/01/01")
	splitTraffic := []int{}
	//smoothing := []int{}

	for ts := 0; ts < numSteps; ts++ {
		fmt.Print(ts)
		fmt.Print("\n")
		//bson for holding logs until persistence to database Mongo option
		logs := []interface{}{}
		//string for holding logs to write to csv
		csv := ""
		//choose traffic based on Poisson distribution with expected trafic as lambda
		traffic := int(poisson.Rand())
		//add variation to the traffic total
		traffic = int((float64(traffic) * ((((math.Sin(((float64(ts) * math.Pi) / 84))) / 4) + 0.5) + 1)) * 0.5)

		if float64(traffic/trafficPerStep) > 2 {
			traffic = int(float64(trafficPerStep) * 2)
		}

		smoothing := splitTraffic
		splitTraffic = []int{}
		for f := 0; f < len(functionChains); f++ {
			//split traffic bsed on function chain ratios
			ratioTraffic := (float64(traffic) / 100 * float64(functionChains[f][0])) + 1

			poissonRatio := distuv.Poisson{Lambda: float64(ratioTraffic), Src: srcArrive}

			//Multplying by sin in order to get peaks and troughs in usuage between day and night pluss offset(2) to have peak during day
			augTraffic := int(ratioTraffic * ((math.Sin((((float64(ts) * math.Pi) / 12) - 2)) / 3) + 0.5))

			//simulating bursty traffic at certain times of day
			if realBursty {
				if ratioTraffic/float64(augTraffic) < 2.0 { //Midday slump (people are busy)

					augTraffic = (int((poissonRatio.Rand() * 0.65) * (rand.Float64() + 0.5)))

				} else if ratioTraffic/float64(augTraffic) > 2.5 {

					augTraffic = augTraffic - (augTraffic / (rand.Intn(10) + 1))

				} else {

					augTraffic = int(float64(augTraffic) * (rand.Float64() + 0.75))
				}

			}

			//smoothing algorithm
			if ts > 1 {
				if augTraffic-smoothing[f] > 0 {
					augTraffic = augTraffic - (diff(augTraffic, smoothing[f]) / 2)
				} else {
					augTraffic = augTraffic + (diff(augTraffic, smoothing[f]) / 2)
				}
			}

			splitTraffic = append(splitTraffic, augTraffic)
			//splitTraffic = append(splitTraffic, int(ratioTraffic))

		}

		//for each function chain execute said functions
		for index, value := range splitTraffic {
			for t := 0; t < value; t++ {
				wg.Add(1)
				go realTraffic(&wg, &logs, &csv, startTime, index)
			}
		}

		//BOTNET

		if attackchoice != 0 {
			//C
			if attackIPchoice == 2 || attackIPchoice == 4 {
				//if timestep is a multiple of the reset value, reset IPs
				if ts%resetVal == 0 {
					botnetIPs = botnetIPs[:0]
					//generated IP addresses for the botnet
					for ip := 0; ip < botnetSlice; ip++ {
						botnetIPs = append(botnetIPs, genIpaddr())

					}
				}
			}

			//CD & BCD
			if attackIPchoice == 6 || attackIPchoice == 7 {
				//if timestep is a multiple of the reset value, reset IPs
				if ts%resetVal == 0 {
					seed := rand.Intn(botnetSize - botnetSlice)
					botnetIPs = botnetPoolIPs[seed : seed+botnetSlice]
				}
			}
			//3
			randAttacknum = rand.Intn(randAttackFactor)
			if botnetSize > 0 && attackchoice != 0 {
				if ts >= attackStart && ts < (attackStart+attackDur) {
					fmt.Print("attack start")
					//2
					expoAttacknum = expoAttacknum * expoAttackfactor
					for b := 0; b < botnetSlice; b++ {
						wg.Add(1)
						go botTraffic(&wg, &logs, &csv, startTime, attackchoice)
					}
				}
			}
		}

		wg.Wait()
		AWSLog += AWSLambda.CalculatePrice()
		AWSLog += AWSLambda.CalculateBasePrice()
		GCFLog += GoogleFunctions.CalculatePrice()
		GCFLog += GoogleFunctions.CalculateBasePrice()
		AzureLog += AzureFunctions.CalculatePrice()
		AzureLog += AzureFunctions.CalculateBasePrice()
		IBMLog += IBMFunctions.CalculatePrice()
		IBMLog += IBMFunctions.CalculateBasePrice()
		//increment log time by time step
		startTime = startTime.Add(time.Millisecond * time.Duration(timeFactor))
		//Split large log list for safer writing
		size := 10000
		j := 0
		for i := 0; i < len(logs); i += size {
			j += size
			if j > len(logs) {
				j = len(logs)
			}

			_, insertErr := collection.InsertMany(ctx, logs[i:j])
			if insertErr != nil {
				log.Fatal(insertErr)
			}
		}

		//fmt.Print(expoAttacknum)
		//fmt.Print("\n")
	}

	f, err := os.Create("AWS")
	check(err)
	AWSLog += AWSLambda.AttackDamage()
	w := bufio.NewWriter(f)
	w.WriteString(detailLog)
	w.WriteString(AWSLog)
	w.Flush()
	f, err = os.Create("GCF")
	check(err)
	GCFLog += GoogleFunctions.AttackDamage()
	w = bufio.NewWriter(f)
	w.WriteString(detailLog)
	w.WriteString(GCFLog)
	w.Flush()
	f, err = os.Create("Azure")
	check(err)
	AzureLog += AzureFunctions.AttackDamage()
	w = bufio.NewWriter(f)
	w.WriteString(detailLog)
	w.WriteString(AzureLog)
	w.Flush()
	f, err = os.Create("IBM")
	check(err)
	IBMLog += IBMFunctions.AttackDamage()
	w = bufio.NewWriter(f)
	w.WriteString(detailLog)
	w.WriteString(IBMLog)
	w.Flush()

	f, _ = os.OpenFile("autorun.log", os.O_APPEND, 0644)
	f.Write([]byte(" " + AWSLambda.AttackDamageFMT() + " " + GoogleFunctions.AttackDamageFMT() + " " + AzureFunctions.AttackDamageFMT() + " " + IBMFunctions.AttackDamageFMT() + "\n"))
	f.Close()
}

func SyntheticSetup(input1 int, input2 int, input3 int, input4 int, input5 float64, input6 float64) {
	// input
	attackchoice = input1
	attackIPchoice = input2
	attackStart = input3
	attackDur = input4

	if attackchoice == 1 {
		constantAttacknum = int(input5)
	} else if attackchoice == 2 {
		expoAttackfactor = input6
		expoAttacknum = input5
	} else if attackchoice == 3 {
		randAttackFactor = int(input5)
	}
	fmt.Print(attackchoice)
	fmt.Print(attackIPchoice)

	//HARD CODED USER PARAMETERS  This usecase is taken from https://aws.amazon.com/blogs/compute/load-testing-a-web-applications-serverless-backend/
	//Settings
	userbaseSize = 1000000
	timestep = 2
	numSteps = 730
	usersPerStep = 1500
	trafficPerStep = 50000
	botnetSize = 1000
	botnetSlice = 100
	detailLog += "attack: " + fmt.Sprint(attackchoice) + " IP spoofing: " + fmt.Sprint(attackIPchoice) + "\n"
	detailLog += "all attack params start. Constant = " + fmt.Sprint(constantAttacknum) + "Exponential = " + fmt.Sprint(expoAttacknum) + "\n"
	detailLog += "all IP params start. C reset = " + fmt.Sprint(resetVal) + "B reset = " + fmt.Sprint(5000) + "\n"
	if attackIPchoice == 5 || attackIPchoice == 6 || attackIPchoice == 7 {
		detailLog += "IP Pooling On" + "\n"
	}
	detailLog += "Userbase = " + fmt.Sprint(userbaseSize) + "\n"
	detailLog += "Timestep = " + fmt.Sprint(timestep) + "\n"
	detailLog += "Number of timesteps = " + fmt.Sprint(numSteps) + "\n"
	detailLog += "Users per step = " + fmt.Sprint(usersPerStep) + "\n"
	detailLog += "traffic per step = " + fmt.Sprint(trafficPerStep) + "\n"
	detailLog += "Botnet pool= " + fmt.Sprint(botnetSize) + "\n"
	detailLog += "Botnets usable at once= " + fmt.Sprint(botnetSlice) + "\n"

	//Create functions being targeted
	functions = append(functions, function.Function{ID: 0, Memory: 128, Runtime: 200, Trigger: "dummy"})
	functions = append(functions, function.Function{ID: 1, Memory: 128, Runtime: 217, Trigger: "dummy"})
	functions = append(functions, function.Function{ID: 1, Memory: 128, Runtime: 217, Trigger: "dummy"})
	functions = append(functions, function.Function{ID: 1, Memory: 128, Runtime: 217, Trigger: "dummy"})
	functions = append(functions, function.Function{ID: 2, Memory: 128, Runtime: 200, Trigger: "dummy"})
	functions = append(functions, function.Function{ID: 3, Memory: 128, Runtime: 200, Trigger: "dummy"})
	functions = append(functions, function.Function{ID: 4, Memory: 128, Runtime: 200, Trigger: "dummy"})
	functions = append(functions, function.Function{ID: 5, Memory: 2048, Runtime: 200, Trigger: "dummy"})
	functions = append(functions, function.Function{ID: 6, Memory: 128, Runtime: 200, Trigger: "dummy"})
	functions = append(functions, function.Function{ID: 7, Memory: 128, Runtime: 200, Trigger: "dummy"})
	functions = append(functions, function.Function{ID: 8, Memory: 128, Runtime: 200, Trigger: "dummy"})

	//define logical order function execute in different scenarios first item is ratio % of scenario (expected load = 1000 post questions, 10000 post answers, 50000 get questions)
	//POST Question
	functionChains = append(functionChains, []int{2, 4, 6, 8})
	//POST Answer
	functionChains = append(functionChains, []int{8, 1, 2, 7})
	functionChains = append(functionChains, []int{8, 1, 3, 7})
	//GET Answer
	functionChains = append(functionChains, []int{0, 0})
	//GET Question
	functionChains = append(functionChains, []int{82, 5})

	detailLog += "Function chains" + fmt.Sprint(functionChains) + "\n"
	//generated IP addresses for the user base
	for ip := 0; ip < userbaseSize; ip++ {
		ipaddresses = append(ipaddresses, genIpaddr())
	}

	//generated IP addresses for the botnet
	if botnetSize > 0 {
		for ip := 0; ip < botnetSize; ip++ {
			botnetPoolIPs = append(botnetPoolIPs, genIpaddr())
		}
		botnetIPs = botnetPoolIPs[0:botnetSlice]
	}
	//set time factor
	switch timestep {
	//seconds
	case 0:
		timeFactor = 1000
	//minutes
	case 1:
		timeFactor = 60000
	//hours
	case 2:
		timeFactor = 3.6e6
	//days
	case 3:
		timeFactor = 8.64e7
	//weeks
	case 4:
		timeFactor = 6.048e8
	//months
	case 5:
		timeFactor = 2.628e9
	default:
		timeFactor = 0

	}
}

// Random IP address generator
func genIpaddr() string {
	ip := fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
	return ip
}

func diff(a, b int) int {
	if a < b {
		return b - a
	}
	return a - b
}

// Real traffic generator
func realTraffic(waitG *sync.WaitGroup, logs *[]interface{}, csv *string, startTime time.Time, index int) {
	defer waitG.Done()
	ipaddress := ipaddresses[rand.Intn(userbaseSize)]
	//fmt.Println(ipaddress)
	dif := rand.Intn(timeFactor)
	timestamp := startTime.Add(time.Millisecond * time.Duration(dif))

	for i, value := range functionChains[index] {
		if i != 0 {
			dif := functions[value].Runtime + uint64(rand.Intn(1000))
			timestamp = timestamp.Add(time.Millisecond * time.Duration(dif))
			AWSLambda.RunFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)
			AWSLambda.RunBaseFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)
			GoogleFunctions.RunFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)
			GoogleFunctions.RunBaseFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)
			AzureFunctions.RunFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)
			AzureFunctions.RunBaseFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)
			IBMFunctions.RunFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)
			IBMFunctions.RunBaseFunction(functions[value].ID, functions[value].Runtime, functions[value].Memory)

			mu.Lock()
			if fileWrite == 1 {
				*csv += fmt.Sprintf("%s,false,%d,%s\n", ipaddress, functions[value].ID, timestamp.Format(time.RFC3339))
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
				*logs = append(*logs, bson.D{{Key: "IP", Value: ipaddress}, {Key: "functioID", Value: functions[value].ID}, {Key: "timestamp", Value: timestamp}, {Key: "bot", Value: false}})
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
