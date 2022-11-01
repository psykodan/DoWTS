package UsageGenerator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UsageGenerator struct {
}

//MongoDB connection for logs
/*
   Connect to my cluster
*/
var client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

var ctx, _ = context.WithTimeout(context.Background(), 6000*time.Second)

/*
Get my collection instance
*/
var collection = client.Database("dowsimlogs").Collection("logs")

// logging to file the cost outputs of each platform
var AWSLog, GCFLog, IBMLog, AzureLog, detailLog string

// Mongo or file write
var fileWrite = 1

func Start(input1 int, input2 int, input3 int, input4 int, input5 float64, input6 float64) {

	//SETUP
	//Mongo error handling
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	SyntheticSetup(input1, input2, input3, input4, input5, input6)
	SyntheticDataGenerator()
	//DatasetUsageGenerator()

}

func Auto(input1 int) {

	//SETUP
	//Mongo error handling
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	rand.Seed(time.Now().UTC().UnixNano())

	input2 := 0
	input3 := rand.Intn(730)
	input4 := rand.Intn(730)

	input5 := 0.0
	if input1 == 2 {
		//input5 = float64(7 + rand.Intn(5))
		input5 = 10
		input3 = rand.Intn(48)
		input4 = rand.Intn(230) + 500
	} else {
		input5 = float64(rand.Intn(3000))
	}
	input6 := 1 + (rand.Float64() / (float64(input4) * 0.1))
	if input6 > 1.01 {
		input6 = 1.01
	}
	if input3+input4 > 730 {
		input4 = 730 - input3
	}

	f, err := os.OpenFile("autorun.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(strconv.Itoa(input1) + " " + strconv.Itoa(input2) + " " + strconv.Itoa(input3) + " " + strconv.Itoa(input4) + " " + fmt.Sprintf("%f", input5) + " " + fmt.Sprintf("%f", input6))); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	f, err = os.OpenFile("simdata.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte("IP,bot,functioID,timestamp\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	SyntheticSetup(input1, input2, input3, input4, input5, input6)
	SyntheticDataGenerator()
	//DatasetUsageGenerator()

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
