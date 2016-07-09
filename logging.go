// Logging
package main

import (
	"os"
	"time"
)

/*Logs client actions or messages
ClientToLog (Client pointer, Text string)
Example: ClientToLog(cl, "Changed area to "+cl.area.Name+".")
Output: [127.0.0.1      ][2016-07-09 12:58:03]Changed area to Courtroom 1.
*/
func ClientToLog(cl *Client, logstr string) {
	if !FileExists("ServerLog") {
		CreateFile("ServerLog")
	}
	logfile, err := os.OpenFile("ServerLog", os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()
	ipstring := cl.IP.String() + "        " //Align the IP so the logs don't look wonky
	timeStamp := time.Now().Format("2006-01-02 15:04:05")
	var fullstring string = "[" + ipstring[:15] + "][" + timeStamp + "]" + logstr + "\r\n"
	logfile.WriteString(fullstring)
}

/*Logs server actions or messages
ServerToLog (Text string)
Example: ServerToLog("Starting server.")
Output: [$HOST          ][2016-07-09 12:51:26]Starting server.
*/
func ServerToLog(logstr string) {
	if !FileExists("ServerLog") {
		CreateFile("ServerLog")
	}
	logfile, err := os.OpenFile("ServerLog", os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()
	timeStamp := time.Now().Format("2006-01-02 15:04:05")
	var fullstring string = "[$HOST          ][" + timeStamp + "]" + logstr + "\r\n"
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
