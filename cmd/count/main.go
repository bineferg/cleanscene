package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func parseEu(eu string) float64 {
	euNum := strings.Replace(eu, "€", "", -1)
	f, _ := strconv.ParseFloat(euNum, 64)
	return f
}
func parseKg(kg string) float64 {
	kgNum := strings.Replace(kg, "kg", "", -1)
	kgNumNoSp := strings.Replace(kgNum, " ", "", -1)
	f, _ := strconv.ParseFloat(kgNumNoSp, 64)
	return f
}
func parseL(L string) float64 {
	LNum := strings.Replace(L, "L", "", -1)
	LNumNS := strings.Replace(LNum, " ", "", -1)
	f, _ := strconv.ParseFloat(LNumNS, 64)
	return f
}
func parseKm(km string) float64 {
	kmNum := strings.Replace(km, "km", "", -1)
	kmNumNS := strings.Replace(kmNum, " ", "", -1)
	f, _ := strconv.ParseFloat(kmNumNS, 64)
	return f
}

func countFlights(flight []string) int {
	dep, arr := flight[0], flight[1]
	if dep == arr {
		return 0
	}
	if parseEu(flight[3]) == 0 || parseKg(flight[4]) == 0 || parseL(flight[5]) == 0 || parseKm(flight[6]) == 0 {
		return 0
	}
	return 1

}

func getTotalNumbers(fileName string) (float64, float64, float64, float64, int) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer file.Close()
	var totalEu, totalCarbon, totalFuel, totalDistance float64
	var totalFlights int

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		arr := strings.Split(scanner.Text(), ",")
		if len(arr) < 7 {
			continue
		}
		euoffset := parseEu(arr[3])
		totalEu = totalEu + euoffset
		carbon := parseKg(arr[4])
		totalCarbon = totalCarbon + carbon
		fuel := parseL(arr[5])
		totalFuel = totalFuel + fuel
		distance := parseKm(arr[6])
		totalDistance = totalDistance + distance
		numFlights := countFlights(arr)
		totalFlights = totalFlights + numFlights

	}
	return totalEu, totalCarbon, totalFuel, totalDistance, totalFlights

}

var dir = flag.String("dir", "", "artist dir")

func main() {
	var files []string
	flag.Parse()
	root := fmt.Sprintf("./done/%s/artist-pages/", *dir)

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, "csv") {
			files = append(files, path)
		}
		return nil
	})
	var totalEU, totalCar, totalFuel, totalDist float64
	var totalFlights int
	for _, file := range files {
		eu, car, fuel, dist, tot := getTotalNumbers(file)
		totalEU = totalEU + eu
		totalCar = totalCar + car
		totalFuel = totalFuel + fuel
		totalDist = totalDist + dist
		totalFlights = totalFlights + tot
	}
	fmt.Printf("Total Carbon Offset: %fEU\n", totalEU)
	fmt.Printf("Total Carbon Output: %fKG\n", totalCar)
	fmt.Printf("Total Amount of Fuel Used: %fL \n", totalFuel)
	fmt.Printf("Total Distance Tradeled: %fKm\n", totalDist)
	fmt.Printf("Total Number of Flights Taken: %d \n", totalFlights)
}
