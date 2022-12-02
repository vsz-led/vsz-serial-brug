package main

import (
	"fmt"
	"go.bug.st/serial"
	"log"
)

const baudrate = 115200

func main() {
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
