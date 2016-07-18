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
	"os"
	"time"
)

var clientlog_filename string = "logs/client.log"
var serverlog_filename string = "logs/server.log"

/*Logs client actions or messages
writeClientLog (Client pointer, Text string)
Example: writeClientLog(cl, "Changed area to "+cl.area.Name+".")
Output: [127.0.0.1      ][2016-07-09 14:40:19 UTC][Feen][Phoenix@Courtroom 1] Changed area to Courtroom 1.
*/
func writeClientLog(cl *Client, logstr string) {
	if !FileExists(clientlog_filename) {
		CreateFile(clientlog_filename)
	}
	logfile, err := os.OpenFile(clientlog_filename, os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer logfile.Close()

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	oocchararea := fmt.Sprintf("[%s][%s@%s]", cl.oocname, cl.getCharacterName(), cl.getAreaName())
	fullstring := fmt.Sprintf("[%-15s][%s]%s%s\n", cl.IP.String(), timestamp, oocchararea, logstr)

	logfile.WriteString(fullstring)
}

/*Logs server actions or messages
writeServerLog (Text string)
Example: writeServerLog("Starting server.")
Output: [2016-07-09 12:51:26]Starting server.
*/
func writeServerLog(logstr string) {
	if !FileExists(serverlog_filename) {
		CreateFile(serverlog_filename)
	}
	logfile, err := os.OpenFile(serverlog_filename, os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer logfile.Close()

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	fullstring := fmt.Sprintf("[%s]%s\n", timestamp, logstr)

	logfile.WriteString(fullstring)
}
