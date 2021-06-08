package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

const ddbURL string = "http://ddb.glidernet.org/download/?j=1"

// OGNDevice represents record in OGN device database
type OGNDevice struct {
	DeviceType    string `json:"device_type"`
	DeviceID      string `json:"device_id"`
	AircraftModel string `json:"aircraft_model"`
	Registration  string `json:"registration"`
	CompNumber    string `json:"cn"`
	Tracked       string `json:"tracked"`
	Identified    string `json:"identified"`
}

type ognDatabase struct {
	Devices []OGNDevice `json:"devices"`
}

type converterArgs struct {
	OutputFile string
}

func parseArgs() converterArgs {
	flag.Parse()
	args := flag.Args()
	return converterArgs{OutputFile: args[0]}
}

func parseDatabase(jsonData []byte) []OGNDevice {
	var err error
	// jsonData, err := ioutil.ReadFile(inputFile)
	// if err != nil {
	// 	panic(err)
	// }
	var db ognDatabase
	err = json.Unmarshal(jsonData, &db)
	if err != nil {
		panic(err)
	}
	return db.Devices
}

func generateXML(devices []OGNDevice) bytes.Buffer {
	buf := bytes.Buffer{}
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	buf.WriteString("<FLARMNET Version=\"008423\">\n")

	for _, dvc := range devices {
		buf.WriteString(fmt.Sprintf("<FLARMDATA FlarmID=\"%s\">\n", dvc.DeviceID))
		buf.WriteString("  <NAME></NAME>\n")
		buf.WriteString(fmt.Sprintf("  <AIRFIELD>%s</AIRFIELD>\n", dvc.Registration))
		buf.WriteString(fmt.Sprintf("  <TYPE>%s</TYPE>\n", dvc.AircraftModel))
		buf.WriteString(fmt.Sprintf("  <REG>%s</REG>\n", dvc.Registration))
		buf.WriteString(fmt.Sprintf("  <COMPID>%s</COMPID>\n", dvc.CompNumber))
		buf.WriteString("  <FREQUENCY></FREQUENCY>\n")
		buf.WriteString("</FLARMDATA>\n")
	}

	buf.WriteString("</FLARMNET>\n")
	return buf
}

func generateLXNAV(devices []OGNDevice) bytes.Buffer {
	buf := bytes.Buffer{}
	xmlbuf := generateXML(devices)
	for _, b := range xmlbuf.Bytes() {
		buf.WriteByte(b + 1)
	}
	return buf
}

func fetchDDB() ([]byte, error) {
	resp, err := http.Get(ddbURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// buf := new(strings.Builder)
	// buf.ReadFrom(resp)
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func main() {
	args := parseArgs()
	flag.Parse()

	fmt.Println("Fetching data...")
	json, err := fetchDDB()
	if err != nil {
		panic(err)
	}

	fmt.Println("Generating LXNAV database")
	db := parseDatabase(json)
	buf := generateLXNAV(db)

	fmt.Println("Writing output:", args.OutputFile)
	err = ioutil.WriteFile(args.OutputFile, buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}
