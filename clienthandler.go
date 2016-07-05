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
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

var next_clientid uint64 = 1
var client_list *ClientList = new(ClientList)

func handleClient(conn net.Conn) {
	var n int
	var err error
	var rawmsg string
	var cmd string

	bufsize := 1024
	buf := make([]byte, bufsize)

	addrstring := conn.RemoteAddr().String()
	ipstring := strings.SplitN(addrstring, ":", 2)[0]

	client := Client{}
	client.clientid = next_clientid
	client.charid = -1
	client.IP = net.ParseIP(ipstring)
	client.conn = conn
	client.area = nil
	client.oocname = ""

	next_clientid += 1
	client_list.addClient(&client)
	client.changeAreaID(config.Defaultarea)

	char_list_pages := loadCharPages(10)
	music_list_pages := loadMusicPages(10)

	client.sendRawMessage("decryptor#34#%")

	for {
		if n, err = conn.Read(buf); err != nil {
			log.Printf("Closed connection from %s.", conn.RemoteAddr().String())
			client.disconnect()
			return
		}
		if n == bufsize {
			// TODO make it detect #% as a message separator
			continue
		}
		rawmsg = string(buf[:n])

		if rawmsg[0] == '#' { // encrypted cmd
			rawmsg = rawmsg[1:]
			splitmsg := strings.Split(rawmsg, "#")[0]
			cmd = decryptMessage(hexToBytes(splitmsg), decrypt_key)
		} else if (rawmsg[0] == '3') || (rawmsg[0] == '4') { // encrypted cmd
			splitmsg := strings.Split(rawmsg, "#")[0]
			cmd = decryptMessage(hexToBytes(splitmsg), decrypt_key)
		} else { // plaintext cmd
			cmd = strings.Split(rawmsg, "#")[0]
		}

		fmt.Println(cmd)
		fmt.Println(rawmsg)

		switch cmd {

		case "HI": // initial handshake
			// client ID
			client.sendRawMessage("ID#" + strconv.FormatUint(client.clientid, 10) +
				"#" + server_version + "#%")
			// current players / limit
			client.sendRawMessage("PN#" + strconv.Itoa(client_list.onlineCharacters()) +
				"#" + strconv.Itoa(config.Slots) + "#%")

		case "CH": // client pings the server
			client.sendRawMessage("CHECK#%")
			// give the client X seconds to send another ping
			conn.SetReadDeadline(time.Now().
				Add(time.Duration(config.Timeout) * time.Second))

		case "askchaa": // asking for char/evi/music list lengths
			client.sendRawMessage("SI#" + strconv.Itoa(len(config.Charlist)) +
				"#0#" + strconv.Itoa(len(config.Musiclist)) + "#%")

		case "askchar2": // send list of characters
			client.sendRawMessage(char_list_pages[0])

		case "AN": // character list
			char_start, _ := strconv.Atoi(strings.Split(rawmsg, "#")[1])
			if (char_start < len(char_list_pages)) && (char_start >= 0) {
				client.sendRawMessage(char_list_pages[char_start])
			} else {
				client.sendRawMessage(music_list_pages[0])
			}

		case "AE": // evidence list
			break

		case "AM": // music list
			music_start, _ := strconv.Atoi(strings.Split(rawmsg, "#")[1])
			if (music_start < len(music_list_pages)) && (music_start >= 0) {
				client.sendRawMessage(music_list_pages[music_start])
			} else {
				client.sendDone()
			}

		case "CC": // select character
			char, _ := strconv.Atoi(strings.Split(rawmsg, "#")[2])
			client.charid = char
			// user selected character
			client.sendRawMessage("PV#" + strconv.FormatUint(client.clientid, 10) +
				"#CID#" + strconv.Itoa(client.charid) + "#%")
			// TODO health
			client.sendRawMessage("HP#1#10#%")
			client.sendRawMessage("HP#2#10#%")

		case "MC": // play music
			if out_msg, err := parseMusic(rawmsg, &client); err != nil {
				continue
			} else {
				client.area.sendRawMessage(out_msg)
			}

		case "MS": // IC message
			if out_msg, err := parseMessageIC(rawmsg, &client); err != nil {
				continue
			} else {
				client.area.sendRawMessage(out_msg)
			}

		case "CT": // OOC message
			if out_msg, err := parseMessageOOC(rawmsg, &client); err != nil {
				continue
			} else if out_msg != "" {
				client.area.sendRawMessage(out_msg)
			}
		}
	}
}

/*
MC = Play Music

Example message:
0                   2
MC#Prelude(T&T).mp3#11#%
   1                   3

1 = full song name
2 = char id
*/
func parseMusic(rawmsg string, client *Client) (string, error) {
	split_msg := strings.Split(rawmsg, "#")

	// check if message even makes sense
	if len(split_msg) != 4 {
		return "", errors.New("Message format is wrong.")
	}

	// prepare variables
	songname := split_msg[1]
	charid, err := strconv.Atoi(split_msg[2])
	if err != nil {
		return "", err
	}

	// check if char id matches client
	if charid != client.charid {
		return "", errors.New("Char ID doesn't match client.")
	}

	// check if song name is in the music list
	found := false
	for _, v := range config.Musiclist {
		if v.Name == songname {
			found = true
			break
		}
	}
	if !found {
		return "", errors.New("This song is not on the list.")
	}

	// message is fine
	ret := fmt.Sprintf("MC#%s#%d#%%", songname, charid)
	return ret, nil
}

/*
MS = IC message

Example message:
0       2               4            6            8    10  12   14  16
MS#chat#damage#Portsman#damaged#text#pro#sfx-stab#1#20#1#0#0#20#0#0#%
   1           3                5        7          9    11  13   15

2  = pre-animation
3  = folder name
4  = animation
5  = message text (<= 256 chars)
6  = position (def, pro, jud, wit, hld, hlp)
7  = sound effect (1 if none)
8  = ???
9  = char id
10 = ???
11 = buttons (1 = hold it, 2 = objection, 3 = take that)
12 = ???
13 = char id
14 = ding (1 is ding)
15 = color (0 = black, 1 = green, 2 = red, 3 = orange, 4 = blue)
*/
func parseMessageIC(rawmsg string, client *Client) (string, error) {
	// TODO validate properly
	split_msg := strings.Split(rawmsg, "#")
	ret := "MS#" + strings.Join(split_msg[1:], "#")

	return ret, nil
}

/*
CT = OOC Message

Example message:
0       2
CT#name#text#%
   1         3

1 = OOC name
2 = message text
*/
func parseMessageOOC(rawmsg string, client *Client) (string, error) {
	var ret string = ""
	split_msg := strings.Split(rawmsg, "#")

	// check message format
	if len(split_msg) != 4 {
		return "", errors.New("Message format is wrong.")
	}

	// prepare variables
	oocname := split_msg[1]
	text := split_msg[2]

	// check if message is too long
	if len(text) > 256 {
		return "", errors.New("Message is too long.")
	}

	// check if client has no name assigned yet
	// or if they're using a reserved name
	if oocname == config.Reservedname {
		return "", errors.New("User tried to use a reserved OOC name.")
	} else if client.oocname == "" {
		client.oocname = oocname
	} else if client.oocname != oocname {
		return "", errors.New("User tried to change their OOC name.")
	}

	// process command if message is a command, else just send the message
	if text[0] == '/' {
		split_cmd := strings.Split(text[1:], " ")

		// prepare variables
		cmd := split_cmd[0]
		args := split_cmd[1:]

		// OOC command handling
		switch strings.ToLower(cmd) {
		case "area":
			cmdArea(client, args)
		default:
			client.sendServerMessageOOC("Invalid command.")
		}
	} else {
		ret = fmt.Sprintf("CT#%s#%s#%%", client.oocname, text)
	}
	return ret, nil
}
