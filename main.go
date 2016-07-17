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
	"net"
	"strconv"
)

func main() {
	loadConfig()

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(config.Port))
	if err != nil {
		log.Fatal("An error occurred starting the listening server.")
		writeServerLog("An error occurred starting the listening server.")
	}
	log.Print("Starting server.")
	writeServerLog("Starting server.")

	if err := ban_list.loadBanlist(); err != nil {
		fmt.Println(err.Error())
		writeServerLog(err.Error())
	}

	poll_list.Polls = make(map[string]*Poll)
	if err := poll_list.loadPolls(); err != nil {
		fmt.Println(err.Error())
		writeServerLog(err.Error())
	}

	calcKeys()

	if config.Advertise {
		go msAdvertiser()
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// error accepting connection
			continue
		}
		log.Printf("Accepted connection from %s.", conn.RemoteAddr().String())
		writeServerLog("Accepted connection from " + conn.RemoteAddr().String())
		go handleClient(conn)
	}
}
