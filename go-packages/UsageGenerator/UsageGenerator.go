package UsageGenerator

import (
	"context"
	"log"
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

//logging to file the cost outputs of each platform
var AWSLog, GCFLog, IBMLog, AzureLog, detailLog string

func Start() {

	//SETUP
	//Mongo error handling
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	SyntheticSetup(1, 0, 48, 12)
	SyntheticDataGenerator()
	//DatasetUsageGenerator()

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
