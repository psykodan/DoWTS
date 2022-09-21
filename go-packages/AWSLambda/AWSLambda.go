package AWSLambda

import (
	"fmt"
	"sync/atomic"
)

type AWSLambda struct {
}

//Package that enables the pricing of functions on AWS Lambda as per pricing guide https://aws.amazon.com/lambda/pricing/

//Pricing variables
//Free function requests
var freeRequests = 1000000

//Free function Compute in GBsec
var freeCompute = 400000.0

//Price in $ per GBsec of computation
var COMPUTE_PRICE = 0.00001667

//Price in $ per function request
var REQUEST_PRICE = 0.0000002

//Price in $ per HTTP request under 300m requests
var API_HTTP = 0.000001

//Price in $ per HTTP request over 300m requests
var API_HTTP300M = 0.0000009

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

//Running totals for individual function totals (Hard coded to keep track of 15 functions - can be expanded)
var functions [15][3]uint64

//Function that updates running totals with function invocation
func RunFunction(id int, runtime uint64, memory uint64) {
	//Atomic incrimentation of requests by 1
	atomic.AddUint64(&TotalRequests, 1)
	//Atomic incrimentation of runtime by function runtime in ms
	atomic.AddUint64(&TotalRuntime, runtime)

	//Calculate compute time in GBsec
	var compute = ((float32(runtime)) / 1000) * ((float32(memory)) / 1024)
	//multiply by 1000000 for storage as int that can be safely atomically stored in running total
	//fmt.Print(compute)
	compute = compute * 1000000
	//fmt.Print(compute)
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
	//Subract free requests from total requests
	totalRequests := int(TotalRequests) - freeRequests
	//Multiply remainder by request price to get cost
	reqPrice := float64(totalRequests) * REQUEST_PRICE

	//Subtract free compute time from total compute time (convert compute time back to GBsec by dividing by 1000000)
	totalCompute := float64(TotalCompute)/1000000 - freeCompute
	//Multiply remainder by compute price to get cost
	compPrice := totalCompute * COMPUTE_PRICE

	//Init HTTP request cost to $0
	apiPrice := 0.0
	//If the total requests after the free requests are subreacted is grater than 300m, use lower price for first 300m then use higher price on remainder
	if totalRequests > 300000000 {
		apiPrice = 300000000.0 * API_HTTP
		apiPrice = apiPrice + float64(totalRequests-300000000)*API_HTTP300M
	} else {
		//If the total requests after the free requests are subreacted is less than 300m, use lower price per request
		apiPrice = float64(totalRequests) * API_HTTP
	}

	//Checks to counteract negative prices that occur when below free limits
	if reqPrice < 0 {
		reqPrice = 0
	}

	if compPrice < 0 {
		compPrice = 0
	}

	if apiPrice < 0 {
		apiPrice = 0
	}

	//Total cost
	var price = compPrice + reqPrice //+ apiPrice
	TotalPrice = price
	out := fmt.Sprintf("Total\t\t\t%f,%d,%f\n", price, TotalRequests, float64(TotalCompute)/1000000)

	return out
}

//Function to calculate base price of function executions on platform
func CalculateBasePrice() string {
	//Subract free requests from base requests
	baseRequests := int(BaseRequests) - freeRequests
	//Multiply remainder by request price to get cost
	reqPrice := float64(BaseRequests) * REQUEST_PRICE

	//Subtract free compute time from base compute time (convert compute time back to GBsec by dividing by 1000000)
	baseCompute := float64(BaseCompute)/1000000 - freeCompute
	//Multiply remainder by compute price to get cost
	compPrice := baseCompute * COMPUTE_PRICE

	//Init HTTP request cost to $0
	apiPrice := 0.0
	//If the base requests after the free requests are subreacted is grater than 300m, use lower price for first 300m then use higher price on remainder
	if baseRequests > 300000000 {
		apiPrice = 300000000.0 * API_HTTP
		apiPrice = apiPrice + float64(baseRequests-300000000)*API_HTTP300M
	} else {
		//If the base requests after the free requests are subreacted is less than 300m, use lower price per request
		apiPrice = float64(baseRequests) * API_HTTP
	}

	//Checks to counteract negative prices that occur when below free limits
	if reqPrice < 0 {
		reqPrice = 0
	}

	if compPrice < 0 {
		compPrice = 0
	}

	if apiPrice < 0 {
		apiPrice = 0
	}

	//Base cost
	var price = compPrice + reqPrice //+ apiPrice
	BasePrice = price
	out := fmt.Sprintf("Base total\t\t%f,%d,%f\n", price, BaseRequests, float64(BaseCompute)/1000000)

	return out
}

func AttackDamage() string {
	damage := TotalPrice - BasePrice

	out := fmt.Sprintf("Damage Caused\t%f\n", damage)

	return out
}

//Function that calculates cost per function. Works the same as total cost function except returns array of prices of each function
func CalculateFnPrice(numfn int) []float64 {
	var prices []float64

	for i := 0; i < numfn; i++ {

		totalRequests := int(functions[i][0]) - freeRequests

		reqPrice := float64(totalRequests) * REQUEST_PRICE

		totalCompute := float64(functions[i][2])/1000000 - freeCompute
		compPrice := totalCompute * COMPUTE_PRICE

		apiPrice := 0.0

		if totalRequests > 300000000 {
			apiPrice = 300000000.0 * API_HTTP
			apiPrice = apiPrice + float64(totalRequests-300000000)*API_HTTP300M
		} else {
			apiPrice = float64(totalRequests) * API_HTTP
		}

		if reqPrice < 0 {
			reqPrice = 0
		}

		if compPrice < 0 {
			compPrice = 0
		}

		if apiPrice < 0 {
			apiPrice = 0
		}

		var price = compPrice + reqPrice + apiPrice

		prices = append(prices, price)
	}
	return prices
}
