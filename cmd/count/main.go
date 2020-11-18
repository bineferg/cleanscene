package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
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

func countAll(files []string) {
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

type stat struct {
	name string
	val  float64
}

func compare(stats []stat, val float64) {
	var compVal float64
	sort.Slice(stats, func(i, j int) bool {
		return int(stats[i].val) < int(stats[j].val)
	})
	var numBottomArtists = 0
	for _, stat := range stats {
		if compVal < val {
			compVal = compVal + stat.val
			numBottomArtists++
		}
	}
	fmt.Printf("Equivalient to the bottom %d number of artsts", numBottomArtists)

}

func countTopN(files []string, field string, n int) {
	var stats = make([]stat, 0)
	for _, file := range files {
		switch field {
		case "offset":
			eu, _, _, _, _ := getTotalNumbers(file)
			stats = append(stats, stat{name: file, val: eu})
		case "carbon":
			_, car, _, _, _ := getTotalNumbers(file)
			stats = append(stats, stat{name: file, val: car})
		case "fuel":
			_, _, fuel, _, _ := getTotalNumbers(file)
			stats = append(stats, stat{name: file, val: fuel})
		case "distance":
			_, _, _, dist, _ := getTotalNumbers(file)
			stats = append(stats, stat{name: file, val: dist})
		case "flights":
			_, _, _, _, flights := getTotalNumbers(file)
			stats = append(stats, stat{name: file, val: float64(flights)})
		}
	}
	sort.Slice(stats, func(i, j int) bool {
		return int(stats[i].val) > int(stats[j].val)
	})
	fmt.Printf("The top %d highest numbers for %s are...\n", n, field)
	var totalVal float64
	for _, stat := range stats[:n] {
		totalVal = totalVal + stat.val
		pathName := strings.Split(stat.name, "/")
		name := strings.Replace(pathName[len(pathName)-1], ".csv", "", -1)
		fmt.Printf("%s: %f\n", name, stat.val)
	}
	fmt.Printf("Totaling at: %f", totalVal)
	compare(stats, totalVal)
}

var param = flag.String("param", "", "what you want to count")
var countN = flag.Int("N", 10, "number of top artists you want the count for")
var field = flag.String("field", "", "what you want the total of")

func main() {
	var files []string
	flag.Parse()
	root := "./output/artist-pages/cleanup"

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, "csv") {
			files = append(files, path)
		}
		return nil
	})

	switch *param {
	case "total":
		countAll(files)
	case "top-N":
		countTopN(files, *field, *countN)
	default:
		log.Fatal("nothing to count :/")
	}
}
