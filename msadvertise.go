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
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	advertiseMS bool
	advertiseMu sync.Mutex
)

func msAdvertiser() {
	var conn net.Conn
	var is_connected bool = false
	var reader *bufio.Reader
	var lock sync.Mutex

	// start pinging the masterserver
	ticker := time.NewTicker(15 * time.Second)
	go func() {
		for range ticker.C {
			advertising := isAdvertising()
			if !advertising {
				return
			}
			lock.Lock()
			if is_connected {
				conn.Write([]byte("PING#%"))
			}
			lock.Unlock()
		}
	}()

	for {
		advertising := isAdvertising()

		// check if MS is connected
		if !is_connected {
			if c, err := msConnect(); err != nil {
				writeServerLog("Failed to connect to master server. Retrying.")
				time.Sleep(10 * time.Second)
				continue
			} else {
				lock.Lock()
				writeServerLog("Connected to master server.")
				is_connected = true
				conn = c
				reader = bufio.NewReader(conn)
				msSendInfo(conn)
				lock.Unlock()
			}
		}
		if !advertising {
			conn.Close()
			return
		}

		// read data
		str, err := reader.ReadString('%')
		if err != nil {
			lock.Lock()
			is_connected = false
			lock.Unlock()
			writeServerLog("Disconnected from master server. Retrying.")
			continue
		}
		switch str {
		case "NOSERV#%":
			msSendInfo(conn)
		}

	}
}

func msConnect() (net.Conn, error) {
	if conn, err := net.Dial("tcp", config.Masterserver); err != nil {
		return conn, err
	} else {
		return conn, nil
	}
}

func msSendInfo(conn net.Conn) {
	_, err := conn.Write([]byte(
		fmt.Sprintf("SCC#%d#%s#%s#%s#%",
			config.Port, config.Servername, config.Description, server_version)))
	if err != nil {
		writeServerLog("Failed to publish server.")
	} else {
		writeServerLog("Server published on master server.")
	}
}

func isAdvertising() bool {
	advertiseMu.Lock()
	defer advertiseMu.Unlock()
	return advertiseMS
}

func setAdvertising(value bool) {
	advertiseMu.Lock()
	advertiseMS = value
	advertiseMu.Unlock()
}
