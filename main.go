package main

import (
    "encoding/csv"
	"strconv"
	"flag"
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"math"
	"time"
	"strings"
	"path/filepath"
)

type CsvLine struct {
    Cups string
	Date string
	Hour int
	KWH float64
	Type string
}


type Config struct {
	Potencia              float64 `json:"Potencia"`
	ImpuestosElectricidad float64 `json:"ImpuestosElectricidad"`
	IVA                   float64 `json:"IVA"`
	Contador              float64 `json:"Contador"`
	Precios               []struct {
		Nombre           string  `json:"Nombre"`
		PrecioPotencia   float64 `json:"PrecioPotencia"`
		PrecioPunta      float64 `json:"PrecioPunta"`
		PrecioValle      float64 `json:"PrecioValle"`
		PrecioSuperValle float64 `json:"PrecioSuperValle"`
		MargenCommercial float64 `json:"MargenCommercial"`
	} `json:"Precios"`
}

func main() {

	csv := flag.String("csv", "./csv/", "Path to csv files")
	configFile := flag.String("config", "./config.json", "Configuration file")
	flag.Parse()

	config := LoadConfig(*configFile)

	filepath.Walk(*csv, func(path string, info os.FileInfo, err error) error {

		if !info.IsDir(){
        	ProcessCSV(config, path)
		}		
        return nil
    })
	
}

func LoadConfig(path string) *Config {

	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	content, _ := ioutil.ReadAll(jsonFile)
	var config Config
	json.Unmarshal(content, &config)

	return &config
}

func ProcessCSV(config *Config, csv string) {

	lines, err := ReadCsv(csv)
    if err != nil {
        panic(err)
	}
	
	totalHourly := make([]float64, 25)

	dateStart := ""
	dateEnd := ""

    // Loop through lines & turn into object
    for _, line := range lines {
		
		h, err1 := strconv.Atoi(line[2])
		f, err2 := strconv.ParseFloat(strings.ReplaceAll(line[3], ",", "."), 64)

		if err1 != nil || err2 != nil {
			continue
		}

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

		if dateStart == "" {
			dateStart = data.Date
		}

		dateEnd = data.Date
	}

	var total float64
	var punta float64
	var valle float64
	var supervalle float64
	
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
	}

	layout := "02/01/2006"
	tStart, err := time.Parse(layout, dateStart)
	tEnd, err := time.Parse(layout, dateEnd)
	days := tEnd.Sub(tStart).Hours() / 24

	fmt.Println("")
	fmt.Println("----------------------------------------------------------------------")
	fmt.Println(dateStart, "->", dateEnd,":", math.Round(total*100)/100, "kWh en", days, "dias ")
	fmt.Println("")	
	fmt.Println("    - Punta :", math.Round(punta*100)/100, "kWh" )
	fmt.Println("    - Valle :", math.Round(valle*100)/100, "kWh" )
	fmt.Println("    - Supervalle :", math.Round(supervalle*100)/100, "kWh" )
	fmt.Println("")

	PrintPrices(config, punta, valle, supervalle, days)

	fmt.Println("")
	fmt.Println("")
}

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

func PrintPrices(config *Config, punta, valle, supervalle, days float64){

	for _, com := range config.Precios {

		p_pot := config.Potencia * days * com.PrecioPotencia
		p_energy := (com.PrecioPunta * punta) + (com.PrecioValle * valle) + (com.PrecioSuperValle * supervalle)
		total := p_pot + p_energy + com.MargenCommercial
		total = total + ( total * (config.ImpuestosElectricidad / 100 ))
		total = total + config.Contador
		total = total + ( total * (config.IVA / 100 ))

		total = math.Round(total*100)/100

		fmt.Println("    ", com.Nombre, ":" , total, "€")
	}
}