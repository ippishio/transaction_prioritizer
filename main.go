package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type transaction struct {
	ID              string
	Amount          float64
	BankCountryCode string
}

var serverDelays = map[string]int{}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
func readCsvFile(filePath string) []transaction {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()
	var Transactions = []transaction{}
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	for index := range records {
		amount, err := strconv.ParseFloat(records[index][1], 64)
		if err != nil {
			log.Fatal("Unable to convert str to float64 ", err)
		} else {
			Transactions = append(Transactions, transaction{ID: records[index][0], Amount: amount, BankCountryCode: records[index][2]})
		}
	}

	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return Transactions
}

func prioritize(transactionss []transaction, totalTime int) ([]transaction, float64) {
	var solution = []transaction{}
	jsonFile, err := os.Open("api_latencies.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &serverDelays)
	var weights, amounts = []int{}, []int{}
	capacity := totalTime
	var transactions = []transaction{}
	for i := range transactionss {
		if capacity >= serverDelays[transactionss[i].BankCountryCode] {
			weights = append(weights, serverDelays[transactionss[i].BankCountryCode])
			amounts = append(amounts, int(transactionss[i].Amount*100))
			transactions = append(transactions, transactionss[i])
		}

	}

	items := len(transactions)
	grid := make([][]int, items+1, items+1)
	grid[0] = make([]int, capacity+1, capacity+1)

	for item := 0; item < items; item++ {

		grid[item+1] = make([]int, capacity+1, capacity+1)
		for k := 0; k < serverDelays[transactions[item].BankCountryCode]; k++ {
			grid[item+1][k] = grid[item][k]
		}
		for k := serverDelays[transactions[item].BankCountryCode]; k <= capacity; k++ {

			grid[item+1][k] = max(grid[item][k], grid[item][k-serverDelays[transactions[item].BankCountryCode]]+int(transactions[item].Amount*100))

		}
	}

	solution_value := grid[items][capacity]
	solution_weight := 0
	var taken []int
	k := capacity
	for item := items; item > 0; item-- {
		if grid[item][k] != grid[item-1][k] {
			taken = append(taken, item-1)
			k -= serverDelays[transactions[item-1].BankCountryCode]
			solution_weight += serverDelays[transactions[item-1].BankCountryCode]
		}
	}
	for index := range taken {
		solution = append(solution, transactions[taken[index]])
	}

	return solution, float64(solution_value) / 100
}
func main() {
	var ms int
	transactions := readCsvFile("transactions.csv")
	fmt.Println("Limit in milliseconds:")
	fmt.Fscan(os.Stdin, &ms)
	solution, value := prioritize(transactions, ms)
	fmt.Println(solution)
	fmt.Println(value)
}
