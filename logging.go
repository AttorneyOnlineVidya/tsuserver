/*
tsuserver, an Attorney Online server
Copyright (C) 2016 tsukasa84 <tsukasadev84@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

var ClientLog *log.Logger
var ServerLog *log.Logger

func initLogging() {
	clientHandle, err := os.OpenFile("logs/client.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
	}

	serverHandle, err := os.OpenFile("logs/server.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
	}

	ClientLog = log.New(clientHandle, "", 0)
	ServerLog = log.New(serverHandle, "", 0)
}

/*Logs client actions or messages
writeClientLog (Client pointer, Text string)
Example: writeClientLog(cl, "Changed area to "+cl.area.Name+".")
Output: [127.0.0.1      ][2016-07-09 14:40:19 UTC][Feen][Phoenix@Courtroom 1] Changed area to Courtroom 1.
*/
func writeClientLog(cl *Client, logstr string) {
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	oocchararea := fmt.Sprintf("[%s][%s@%s]", cl.oocname, cl.getCharacterName(), cl.getAreaName())
	fullstring := fmt.Sprintf("[%-15s][%s]%s%s", cl.IP.String(), timestamp, oocchararea, logstr)

	ClientLog.Println(fullstring)
}

/*Logs server actions or messages
writeServerLog (Text string)
Example: writeServerLog("Starting server.")
Output: [2016-07-09 12:51:26]Starting server.
*/
func writeServerLog(logstr string) {
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	fullstring := fmt.Sprintf("[%s]%s", timestamp, logstr)

	ServerLog.Println(fullstring)
}
