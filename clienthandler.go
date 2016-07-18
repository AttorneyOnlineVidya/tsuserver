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

	"golang.org/x/time/rate"
)

var next_clientid uint64 = 1

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
	client.HDID = ""
	client.conn = conn
	client.area = nil
	client.oocname = ""
	client.is_mod = false
	client.muted = true
	client.pos = ""
	client.advert = true
	client.global = true
	client.dj = true

	next_clientid += 1
	client_list.addClient(&client)
	client.changeAreaID(config.Defaultarea)

	char_list_pages := loadCharPages(10)
	music_list_pages := loadMusicPages(10)

	// 5 messages per second, max burst of 8
	rate_limiter := rate.NewLimiter(5, 8)
	spamming := false

	client.sendRawMessage("decryptor#34#%")

	for {
		if isValidCharID(client.charid) && !rate_limiter.Allow() {
			spamming = true
			continue
		}

		if spamming {
			spamming = false
			client.sendServerMessageOOC("Stop spamming the game!")
			writeServerLog(fmt.Sprintf("Client spamming. IP: %s, HD: %s.",
				client.IP, client.HDID))
		}

		if n, err = conn.Read(buf); err != nil {
			log.Printf("Closed connection from %s.", conn.RemoteAddr().String())
			writeServerLog("Closed connection from " + conn.RemoteAddr().String())
			client.disconnect()
			return
		}
		if n == bufsize {
			// TODO make it detect #% as a message separator
			continue
		}
		// message is too short
		if n < 2 {
			continue
		}
		rawmsg = string(buf[:n])

		if rawmsg[0] == '#' { // encrypted cmd
			rawmsg = rawmsg[1:]
			splitmsg := strings.Split(rawmsg, "#")[0]
			cmd = decryptMessage(hexToBytes(splitmsg), crypt_key)
		} else if (rawmsg[0] == '3') || (rawmsg[0] == '4') { // encrypted cmd
			splitmsg := strings.Split(rawmsg, "#")[0]
			cmd = decryptMessage(hexToBytes(splitmsg), crypt_key)
		} else { // plaintext cmd
			cmd = strings.Split(rawmsg, "#")[0]
		}

		fmt.Println(cmd)
		fmt.Println(rawmsg)

		switch cmd {

		case "HI": // initial handshake
			split_msg := strings.Split(rawmsg, "#")
			if len(split_msg) != 3 {
				continue
			}

			// assign HDID
			client.HDID = split_msg[1]

			// check for ban
			if b, _, _ := ban_list.isBanned(&client); b != nil {
				ban_list.addBan(&client, "")
				client.disconnect()
				return
			}

			client.muted = false

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
			continue

		case "AM": // music list
			music_start, _ := strconv.Atoi(strings.Split(rawmsg, "#")[1])
			if (music_start < len(music_list_pages)) && (music_start >= 0) {
				client.sendRawMessage(music_list_pages[music_start])
			} else {
				client.sendDone()
			}

		case "CC": // select character
			split_msg := strings.Split(rawmsg, "#")
			if len(split_msg) != 5 {
				continue
			}
			if char, err := strconv.Atoi(split_msg[2]); err == nil {
				client.changeCharacterID(char)
				client.sendServerMessageOOC(client.getPrintableAreaList())
				cmdMOTD(&client, "")
			}

		case "HP": // penalties
			split_msg := strings.Split(rawmsg, "#")
			if len(split_msg) != 4 {
				continue
			}
			if val, err := strconv.Atoi(split_msg[2]); err == nil {
				if split_msg[1] == "1" {
					if err := client.area.setDefHP(val); err == nil {
						client.area.sendRawMessage(fmt.Sprintf("HP#1#%d#%%", val))
						writeClientLog(&client, "changed health bar")

					}
				} else if split_msg[1] == "2" {
					if err := client.area.setProHP(val); err == nil {
						client.area.sendRawMessage(fmt.Sprintf("HP#2#%d#%%", val))
						writeClientLog(&client, "changed health bar")
					}
				}
			}

		case "opMUTE": // /mute with guard
			split_msg := strings.Split(rawmsg, "#")
			if len(split_msg) != 3 {
				continue
			}
			cmdMute(&client, split_msg[1])

		case "opunMUTE": // /unmute with guard
			split_msg := strings.Split(rawmsg, "#")
			if len(split_msg) != 3 {
				continue
			}
			cmdUnmute(&client, split_msg[1])

		case "RT": // WT/CE buttons
			split_msg := strings.Split(rawmsg, "#")
			if len(split_msg) != 3 {
				continue
			}
			if split_msg[1] == "testimony1" || split_msg[1] == "testimony2" {
				client.area.sendRawMessage(fmt.Sprintf("RT#%s#%", split_msg[1]))
				writeClientLog(&client, "Used WT/CE")
			}

		case "MC": // play music
			if err := parseMusic(rawmsg, &client); err != nil {
				continue
			}

		case "MS": // IC message
			parseMessageIC(rawmsg, &client)

		case "CT": // OOC message
			if out_msg, err := parseMessageOOC(rawmsg, &client); err != nil {
				continue
			} else if out_msg != "" {
				client.area.sendRawMessage(out_msg)
			}

		case "ZZ": // mod call
			client_list.sendAllRawIf(fmt.Sprintf(
				"ZZ#%s (%s) in %s#%%", client.getCharacterName(), client.IP.String(), client.getAreaName()),
				func(c *Client) bool {
					return c.is_mod
				})
			writeClientLog(&client, "used Call Mod")
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
func parseMusic(rawmsg string, client *Client) error {
	split_msg := strings.Split(rawmsg, "#")
	var duration int

	// check if client is muted
	if client.muted {
		return errors.New("Cannot play music, client is muted.")
	}
	// check if client is unDJ'd
	if !client.dj {
		return errors.New("Cannot play music, client is unDJ'd.")
	}

	// check message format
	if len(split_msg) != 4 {
		return errors.New("Message format is wrong.")
	}

	// prepare variables
	songname := split_msg[1]
	charid, err := strconv.Atoi(split_msg[2])
	if err != nil {
		return err
	}

	// check for empty songname
	if len(songname) == 0 {
		return errors.New("Empty song name.")
	}

	// check if char id matches client
	if charid != client.charid {
		return errors.New("Char ID doesn't match client.")
	}
	// Checks if using musiclist to change areas
	if songname[0] == '>' {
		for _, a := range config.Arealist {
			if strings.ToLower(songname[1:len(songname)]) == strings.ToLower(a.Name) {
				if err := client.changeAreaID(a.Areaid); err != nil {
					client.sendServerMessageOOC(err.Error())
					return err
				} else {
					client.sendServerMessageOOC("Changed area to " + client.area.Name + ".")
					writeClientLog(client, "Changed area to "+client.area.Name+".")
					return nil
				}
			}
		}
		client.sendServerMessageOOC("That area could not be found.")
		return nil
	} else {
		// check if song name is in the music list
		found := false
		for _, v := range config.Musiclist {
			if v.Name == songname {
				found = true
				duration = v.Duration
				break
			}
		}
		if !found {
			return errors.New("This song is not on the list.")
		}
	}

	// message is fine
	client.area.playMusic(songname, charid, duration)
	writeClientLog(client, " changed music to "+songname)
	return nil
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
8  = animation type (0 = no pre, 1 = pre, 2 = pre button 5 = zoom)
9  = char id
10 = sound effect delay (>= 0)
11 = buttons (1 = hold it, 2 = objection, 3 = take that)
12 = ???
13 = char id
14 = ding (1 is ding)
15 = color (0 = black, 1 = green, 2 = red, 3 = orange, 4 = blue)
*/
func parseMessageIC(rawmsg string, client *Client) (string, error) {
	split_msg := strings.Split(rawmsg, "#")

	// check if client is muted
	if client.muted {
		return "", errors.New("Cannot send message, client is muted.")
	}

	// check message format
	if len(split_msg) != 17 {
		return "", errors.New("Message format is wrong.")
	}

	// prepare variables
	msgtype := split_msg[1]
	preanim := split_msg[2]
	foldername := split_msg[3]
	anim := split_msg[4]
	text := split_msg[5]
	pos := split_msg[6]
	sfx := split_msg[7]
	animtype := split_msg[8]
	charid1 := split_msg[9]
	sfxdelay := split_msg[10]
	button := split_msg[11]
	unk := split_msg[12]
	charid2 := split_msg[13]
	ding := split_msg[14]
	color := split_msg[15]

	// check msgtype
	if msgtype != "chat" {
		return "", errors.New("Invalid message type.")
	}

	// check text length and trim if needed
	if len(text) > 256 {
		text = text[:256]
	}

	// check if user has a custom pos, else check if pos is valid
	userpos := client.getPosition()
	if userpos != "" {
		pos = userpos
	} else {
		if !isPosValid(pos) {
			return "", errors.New("Invalid position.")
		}
	}

	// check if animtype is valid
	if at, err := strconv.Atoi(animtype); err != nil {
		return "", errors.New("Animation type must be a number.")
	} else {
		if !(at >= 0 && at <= 2) && !(at >= 5 && at <= 6) {
			return "", errors.New("Invalid animation type.")
		}
	}

	// check char ids
	if cid1, err := strconv.Atoi(charid1); err != nil {
		return "", errors.New("Character ID 1 must be an integer.")
	} else {
		if cid2, err := strconv.Atoi(charid2); err != nil {
			return "", errors.New("Character ID 2 must be an integer..")
		} else {
			if cid1 != client.charid {
				return "", errors.New("Character ID different from client.")
			} else if cid1 != cid2 {
				return "", errors.New("Character IDs don't match.")
			} else if !isValidCharID(cid1) {
				return "", errors.New("Character ID is invalid.")
			}
		}
	}

	// check sfx delay
	if del, err := strconv.Atoi(sfxdelay); err != nil {
		return "", errors.New("SFX Delay must be a number.")
	} else {
		if del < 0 {
			return "", errors.New("SFX Delay must be a positive number.")
		}
	}

	// check button
	if but, err := strconv.Atoi(button); err != nil {
		return "", errors.New("Button ID must be a number.")
	} else {
		if !(but >= 0 && but <= 3) {
			return "", errors.New("Invalid button ID.")
		}
	}

	// check ding
	if d, err := strconv.Atoi(ding); err != nil {
		return "", errors.New("Ding must be a number.")
	} else {
		if d != 0 && d != 1 {
			return "", errors.New("Invalid ding ID.")
		}
	}

	// check color
	if col, err := strconv.Atoi(color); err != nil {
		return "", errors.New("Color must be a number.")
	} else if col == 2 && client.is_mod == false { //Reserves redtext for mod only
		color = "0"
	} else {
		if !(col >= 0 && col <= 4) {
			return "", errors.New("Invalid color ID.")
		}
	}

	// return message
	ret := fmt.Sprintf("MS#%s#%s#%s#%s#%s#%s#%s#%s#%d#%s#%s#%s#%d#%s#%s#%%",
		msgtype, preanim, foldername, anim, text, pos, sfx, animtype,
		client.charid, sfxdelay, button, unk, client.charid, ding, color)

	if client.getAreaPtr().sendICMessage(ret, len(text)) {
		writeClientLog(client, "[IC] "+text)
	}

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
	if isOOCNameReserved(oocname) {
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
		target := ""

		if len(args) > 0 {
			target = strings.Join(args, " ")
		}

		// OOC command handling
		switch strings.ToLower(cmd) {
		case "area":
			cmdArea(client, args)
		case "getarea":
			cmdGetArea(client, target)
		case "getareas":
			cmdGetAllAreas(client)
		case "login":
			cmdLogin(client, args)
		case "iplist":
			cmdIpList(client)
		case "mute":
			cmdMute(client, target)
		case "unmute":
			cmdUnmute(client, target)
		case "kick":
			cmdKick(client, target)
		case "ban":
			cmdBan(client, args)
		case "reloadbans":
			cmdReloadBans(client)
		case "bg":
			cmdBackground(client, args)
		case "bglock":
			cmdBgLock(client)
		case "switch":
			cmdSwitch(client, target)
		case "charselect":
			cmdCharselect(client, target)
		case "randomchar":
			cmdRandomChar(client)
		case "pm":
			cmdPM(client, target)
		case "pos":
			cmdPos(client, target)
		case "g":
			cmdGlobalMessage(client, target)
		case "global":
			cmdGlobalToggle(client)
		case "need":
			cmdNeed(client, target)
		case "adverts":
			cmdAdvertToggle(client)
		case "announce":
			cmdModAnnounce(client, target)
		case "motd":
			cmdMOTD(client, target)
		case "roll":
			cmdRoll(client, target)
		case "help":
			cmdHelp(client)
		case "status":
			cmdStatus(client, target)
		case "lm":
			cmdLocalMod(client, target)
		case "gm":
			cmdGlobalMod(client, target)
		case "setdoc":
			cmdSetDoc(client, target)
		case "doc":
			cmdGetDoc(client)
		case "newpoll":
			cmdNewPoll(client, target)
		case "pollresults":
			cmdPollResults(client, target)
		case "closepoll":
			cmdClosePoll(client, target)
		case "vote":
			cmdVote(client, args)
		case "polls":
			cmdPolls(client)
		case "reloadpolls":
			cmdReloadPolls(client)
		case "dj":
			cmdDJ(client, target)
		case "undj":
			cmdUnDJ(client, target)
		default:
			client.sendServerMessageOOC("Invalid command.")
		}
	} else {
		ret = fmt.Sprintf("CT#%s#%s#%%", client.oocname, text)
		writeClientLog(client, "[OOC] "+text) //OOC chat will have [OOC] in log
	}
	return ret, nil
}
