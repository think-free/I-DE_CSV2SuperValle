package main

import (
    "encoding/csv"
	"strconv"
	"flag"
	"os"
	"fmt"
	"strings"
)

type CsvLine struct {
    Cups string
	Date string
	Hour int
	KWH float64
	Type string
}

func main() {

	csv := flag.String("csv", "consumo.csv", "El archivo csv de i-de")
	flag.Parse()

    lines, err := ReadCsv(*csv)
    if err != nil {
        panic(err)
	}
	
	totalHourly := make([]float64, 25)

    // Loop through lines & turn into object
    for _, line := range lines {
		
		h, _ := strconv.Atoi(line[2])
		f, _ := strconv.ParseFloat(strings.ReplaceAll(line[3], ",", "."), 64)

        data := CsvLine{
            Cups: line[0],
			Date: line[1],
			Hour: h,
			KWH: f,
			Type: line[4],
		}
		if data.Hour < 25 { // Yeah, I know but had a csv with hour == 25 O_o
			totalHourly[data.Hour] = totalHourly[data.Hour] + data.KWH
		}		
	}

	var total float64
	var punta float64
	var valle float64
	var supervalle float64
	var sum float64
	
	for i, v := range totalHourly{

		if i == 0 {
			// Skipping index 0
			continue

		} else if i == 1 {
			valle = valle + v
		} else if i < 8 {
			supervalle = supervalle + v
		} else if i < 14 {
			valle = valle + v
		} else if i < 24 {
			punta = punta + v
		} else {
			valle = valle + v
		}

		total = total + v

		fmt.Println("Hour :", i-1, "->" ,i, "kWh :", v)
	}

	sum = punta + valle + supervalle

	fmt.Println("")
	fmt.Println("-----------------------------------")
	fmt.Println("Punta :", punta)
	fmt.Println("Valle :", valle)
	fmt.Println("Supervalle :", supervalle)
	fmt.Println("-----------------------------------")
	fmt.Println("Total :", total, "(" , sum , ")" )
}

// ReadCsv accepts a file and returns its content as a multi-dimentional type
// with lines and each column. Only parses to string type.
func ReadCsv(filename string) ([][]string, error) {

    // Open CSV file
    f, err := os.Open(filename)
    if err != nil {
        return [][]string{}, err
    }
    defer f.Close()

	// Read File into a Variable
	r := csv.NewReader(f)
	r.Comma = ';'
    lines, err := r.ReadAll()
    if err != nil {
        return [][]string{}, err
    }

    return lines, nil
}