package IBMFunctions

import (
	"fmt"
	"sync/atomic"
)

type IBMFunctions struct {
}

//Package that enables the pricing of functions on IBM Cloud Functions as per pricing guide https://cloud.ibm.com/functions/learn/pricing

//Pricing variables
//Free function Compute in GBsec
var freeCompute = 400000.0

//Price in $ per GBsec of computation
var COMPUTE_PRICE = 0.000017

//Running totals of total requests, function runtime and function computation on platform
var TotalRequests uint64
var TotalRuntime uint64
var TotalCompute uint64
var TotalPrice = 0.0

//Running totals of just non malicious requests, for purpose of calculating attack damage
var BaseRequests uint64
var BaseRuntime uint64
var BaseCompute uint64
var BasePrice = 0.0

//Running totals for individual function totals (Hard coded to keep track of 10 functions - can be expanded)
var functions [10][3]uint64

//Function that updates running totals with function invocation
func RunFunction(id int, runtime uint64, memory uint64) {
	//Atomic incrimentation of requests by 1
	atomic.AddUint64(&TotalRequests, 1)
	//Atomic incrimentation of runtime by function runtime in ms
	atomic.AddUint64(&TotalRuntime, runtime)

	//Calculate compute time in GBsec
	var compute = ((float32(runtime)) / 1000) * ((float32(memory)) / 1024)
	//multiply by 1000000 for storage as int that can be safely atomically stored in running total
	compute = compute * 1000000
	atomic.AddUint64(&TotalCompute, uint64(compute))

	//Use function id (0,1,2...,n) to store seperate function totals
	atomic.AddUint64(&functions[id][0], 1)
	atomic.AddUint64(&functions[id][1], runtime)
	atomic.AddUint64(&functions[id][2], uint64(compute))

}

//Function that updates running totals with function invocation
func RunBaseFunction(id int, runtime uint64, memory uint64) {
	//Atomic incrimentation of requests by 1
	atomic.AddUint64(&BaseRequests, 1)
	//Atomic incrimentation of runtime by function runtime in ms
	atomic.AddUint64(&BaseRuntime, runtime)

	//Calculate compute time in GBsec
	var compute = ((float32(runtime)) / 1000) * ((float32(memory)) / 1024)
	//multiply by 1000000 for storage as int that can be safely atomically stored in running total
	//fmt.Print(compute)
	compute = compute * 1000000
	//fmt.Print(compute)
	atomic.AddUint64(&BaseCompute, uint64(compute))

	//Use function id (0,1,2...,n) to store seperate function totals
	atomic.AddUint64(&functions[id][0], 1)
	atomic.AddUint64(&functions[id][1], runtime)
	atomic.AddUint64(&functions[id][2], uint64(compute))

}

//Function to calculate total price of function executions on platform
func CalculatePrice() string {

	//Subtract free compute time from total compute time (convert compute time back to GBsec by dividing by 1000000)
	totalCompute := float64(TotalCompute)/1000000 - freeCompute
	//Multiply remainder by compute price to get cost
	compPrice := totalCompute * COMPUTE_PRICE

	//Checks to counteract negative prices that occur when below free limits
	if compPrice < 0 {
		compPrice = 0
	}
	//Total cost
	var price = compPrice

	out := fmt.Sprintf("Total\t\t\t%f,%d,%f\n", price, TotalRequests, float64(TotalCompute)/1000000)
	return out
}

//Function to calculate Base price of function executions on platform
func CalculateBasePrice() string {

	//Subtract free compute time from Base compute time (convert compute time back to GBsec by dividing by 1000000)
	baseCompute := float64(BaseCompute)/1000000 - freeCompute
	//Multiply remainder by compute price to get cost
	compPrice := baseCompute * COMPUTE_PRICE

	//Checks to counteract negative prices that occur when below free limits
	if compPrice < 0 {
		compPrice = 0
	}
	//Total cost
	var price = compPrice

	out := fmt.Sprintf("Base total\t\t%f,%d,%f\n", price, BaseRequests, float64(BaseCompute)/1000000)
	return out
}

//Function that calculates cost per function. Works the same as total cost function except returns array of prices of each function
func CalculateFnPrice(numfn int) []float64 {
	var prices []float64

	for i := 0; i < numfn; i++ {

		totalCompute := float64(functions[i][2])/1000000 - freeCompute
		compPrice := totalCompute * COMPUTE_PRICE

		if compPrice < 0 {
			compPrice = 0
		}

		var price = compPrice
		//fmt.Printf("Total Cost - $ %f \nNumber of requests - %d \nTotal Compute - %fGBs", price, TotalRequests, float64(TotalCompute)/1000000)
		prices = append(prices, price)
	}
	return prices
}

func AttackDamage() string {
	damage := TotalPrice - BasePrice

	out := fmt.Sprintf("Damage Caused\t%f\n", damage)

	return out
}
