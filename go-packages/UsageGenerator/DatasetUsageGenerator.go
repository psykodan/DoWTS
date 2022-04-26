package UsageGenerator

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dowts/go-packages/AWSLambda"
	"github.com/dowts/go-packages/AzureFunctions"
	"github.com/dowts/go-packages/GoogleFunctions"
	"github.com/dowts/go-packages/IBMFunctions"
	"github.com/dowts/go-packages/function"
)

//Create functions based on url enpoints in dataset
var fn_triggers = make(map[string]function.Function)

//logging to file the cost outputs of each platform
//var AWSLog, GCFLog, IBMLog, AzureLog string

func DatasetUsageGenerator() {

	check := func(e error) {
		if e != nil {
			panic(e)
		}
	}

	// open dataset file
	f, err := os.Open("../eCommerce Events History in Cosmetics Shop/2020-Jan.csv")
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader in order to populate functions
	csvReader := csv.NewReader(f)
	csvReader.Read()
	for {

		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		//add functions to set with generic id
		fn_triggers[rec[1]] = function.Function{ID: 0, Memory: 128, Runtime: 500, Trigger: rec[1]}

	}
	//assign ID to functions
	fn_ID := 0
	for t := range fn_triggers {
		fn_triggers[t] = function.Function{ID: fn_ID, Memory: 128, Runtime: 500, Trigger: t}
		fn_ID++
	}

	fmt.Print(fn_triggers)

	//rewind to start of csv file
	f.Seek(0, io.SeekStart)

	//execute function as each enpoint triggered in dataset
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		AWSLambda.RunFunction(fn_triggers[rec[1]].ID, fn_triggers[rec[1]].Runtime, fn_triggers[rec[1]].Memory)
		GoogleFunctions.RunFunction(fn_triggers[rec[1]].ID, fn_triggers[rec[1]].Runtime, fn_triggers[rec[1]].Memory)
		AzureFunctions.RunFunction(fn_triggers[rec[1]].ID, fn_triggers[rec[1]].Runtime, fn_triggers[rec[1]].Memory)
		IBMFunctions.RunFunction(fn_triggers[rec[1]].ID, fn_triggers[rec[1]].Runtime, fn_triggers[rec[1]].Memory)
	}

	//Calculate price
	AWSLog += AWSLambda.CalculatePrice()
	GCFLog += GoogleFunctions.CalculatePrice()
	AzureLog += AzureFunctions.CalculatePrice()
	IBMLog += IBMFunctions.CalculatePrice()

	//Write to file
	out, err := os.Create("AWS")
	check(err)
	w := bufio.NewWriter(out)
	w.WriteString(AWSLog)
	w.Flush()
	out, err = os.Create("GCF")
	check(err)
	w = bufio.NewWriter(out)
	w.WriteString(GCFLog)
	w.Flush()
	out, err = os.Create("Azure")
	check(err)
	w = bufio.NewWriter(out)
	w.WriteString(AzureLog)
	w.Flush()
	out, err = os.Create("IBM")
	check(err)
	w = bufio.NewWriter(out)
	w.WriteString(IBMLog)
	w.Flush()
}
