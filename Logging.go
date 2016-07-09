// Logging
package main

import (
	"os"
	"time"
)

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
	var fullstring string = "[" + ipstring[:15] + "][" + time.Now().Format(time.RFC1123) + "][" + cl.getCharacterName() + "]" + logstr + "\r\n"
	logfile.WriteString(fullstring)
}

func ServerToLog(logstr string) {
	if !FileExists("ServerLog") {
		CreateFile("ServerLog")
	}
	logfile, err := os.OpenFile("ServerLog", os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()
	var fullstring string = "[$HOST          ][" + time.Now().Format(time.RFC1123) + "]" + logstr + "\r\n"
	logfile.WriteString(fullstring)
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func CreateFile(name string) error {
	fo, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		fo.Close()
	}()
	return nil
}
