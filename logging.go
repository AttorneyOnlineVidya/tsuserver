// Logging
package main

import (
	"fmt"
	"os"
	"time"
)

var clientlog_filename string = "logs/client.log"
var serverlog_filename string = "logs/server.log"

/*Logs client actions or messages
ClientToLog (Client pointer, Text string)
Example: ClientToLog(cl, "Changed area to "+cl.area.Name+".")
Output: [127.0.0.1      ][2016-07-09 12:58:03]Changed area to Courtroom 1.
*/
func ClientToLog(cl *Client, logstr string) {
	if !FileExists(clientlog_filename) {
		CreateFile(clientlog_filename)
	}
	logfile, err := os.OpenFile(clientlog_filename, os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	fullstring := fmt.Sprintf("[%-15s][%s]%s\n", cl.IP.String(), timestamp, logstr)

	logfile.WriteString(fullstring)
}

/*Logs server actions or messages
ServerToLog (Text string)
Example: ServerToLog("Starting server.")
Output: [2016-07-09 12:51:26]Starting server.
*/
func ServerToLog(logstr string) {
	if !FileExists(serverlog_filename) {
		CreateFile(serverlog_filename)
	}
	logfile, err := os.OpenFile(serverlog_filename, os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	fullstring := fmt.Sprintf("[%s]%s\n", timestamp, logstr)

	logfile.WriteString(fullstring)
}

//Checks if the log file exists
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

//Creates the log file
func CreateFile(name string) error {
	fo, err := os.Create(name)
	if err != nil {
		return err
	}
	defer fo.Close()
	return nil
}
