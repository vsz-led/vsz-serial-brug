package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"go.bug.st/serial"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"time"
)

type config struct {
	Mysql struct {
		Host string `yaml:"host"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
		DB   string `yaml:"db"`
	}
	Kruisingscode int
}

type kruising struct {
	Plaats       string
	Weg          string
	Bedrijfsnaam string
}

const baudrate = 115200

func main() {
	// Read config
	file, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Printf("failed to read config: %s", err)
		return
	}

	// Create config object
	var config config

	// Parse config to object
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Printf("failed to parse config: %s", err)
		return
	}

	// Retrieve the port list
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Printf("Failed to get ports list: %s", err)
		return
	}

	// Check amount of serial ports
	if len(ports) == 0 {
		log.Println("No serial ports found!")
		return
	}

	// Print the list of detected ports
	for _, port := range ports {
		log.Printf("Found serial port: %s\n", port)
	}

	// Configure the serial port
	mode := &serial.Mode{
		BaudRate: baudrate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	// Setup mysql connection
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", config.Mysql.User, config.Mysql.Pass, config.Mysql.Host, config.Mysql.DB))
	if err != nil {
		log.Printf("failed to start mysql connection: %s", err)
		return
	}
	// Configure mysql connection
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	rows, err := db.Query("SELECT plaats,weg,bedrijfsnaam FROM Kruising JOIN Opdrachtgever ON Kruising.bedrijfscode=Opdrachtgever.bedrijfscode WHERE kruisingscode = ?", config.Kruisingscode)
	if err != nil {
		log.Printf("Failed to query: %s", err)
		return
	}

	succes := false
	for rows.Next() {
		var kruising kruising
		err = rows.Scan(&kruising.Plaats, &kruising.Weg, &kruising.Bedrijfsnaam)
		if err != nil {
			log.Printf("Failed to scan query: %s", err)
			return
		}

		succes = true
		log.Printf("Starting kruising %s in %s for opdrachtgever %s", kruising.Weg, kruising.Plaats, kruising.Bedrijfsnaam)
	}

	// Check if there was a row, if not shutdown
	if !succes {
		log.Printf("Failed to find kruising with code %d, shutting down", config.Kruisingscode)
		return
	}

	// Open the serial port
	port, err := serial.Open(ports[0], mode)
	if err != nil {
		log.Printf("Failed to open port %s: %s", port, err)
		return
	}

	// Create buffer for reading serial with size 100
	buff := make([]byte, 100)

	for {
		// Read from serial into buffer
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal(err)
		}

		// Check if the length is 0
		if n == 0 {
			fmt.Println("\nEOF")
			break
		}

		// Put received data into string
		data := string(buff[:n])

		// Print received data
		fmt.Print(data)
	}
}
