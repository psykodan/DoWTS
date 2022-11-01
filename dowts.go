package main

import (
	"os"
	"strconv"

	"github.com/dowts/go-packages/UsageGenerator"
)

func main() {
	in1, _ := strconv.Atoi(os.Args[1])
	//in2, _ := strconv.Atoi(os.Args[2])
	//in3, _ := strconv.Atoi(os.Args[3])
	//in4, _ := strconv.Atoi(os.Args[4])
	//in5, _ := strconv.ParseFloat(os.Args[5], 64)
	//in6, _ := strconv.ParseFloat(os.Args[6], 64)

	//UsageGenerator.Start(in1, in2, in3, in4, in5, in6)
	UsageGenerator.Auto(in1)
}
