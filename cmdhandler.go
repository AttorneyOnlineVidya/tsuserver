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
	"strconv"
	"strings"
)

func cmdArea(cl *Client, args []string) {
	if len(args) == 0 {
		cl.sendServerMessageOOC(cl.getPrintableAreaList())
	} else if len(args) == 1 {
		targetarea, err := strconv.Atoi(args[0])
		if err != nil {
			cl.sendServerMessageOOC("The argument must be a number.")
			return
		}
		if err := cl.changeAreaID(targetarea); err != nil {
			cl.sendServerMessageOOC(err.Error())
		} else {
			cl.sendServerMessageOOC("Changed area to " + cl.area.Name + ".")
			writeClientLog(cl, "Changed area to "+cl.area.Name+".")
		}
	} else {
		cl.sendServerMessageOOC("Too many arguments.")
	}
}

func cmdGetArea(cl *Client, id string) {
	var aptr *Area
	ret := "\r\n"

	if len(id) == 0 {
		// print current area
		aptr = cl.getAreaPtr()
	} else {
		// print target area
		if areaid, err := strconv.Atoi(id); err == nil {
			aptr = getAreaPtr(areaid)
		} else {
			cl.sendServerMessageOOC("Argument must be a number.")
			return
		}
	}
	if aptr == nil {
		cl.sendServerMessageOOC("Invalid area ID.")
		return
	}

	ret += fmt.Sprintf("=== Area %d: %s ===", aptr.Areaid, aptr.Name)
	for _, c := range aptr.sortedClientsByName() {
		if cl.is_mod {
			ret += fmt.Sprintf("\r\n%s; IP: %s", c.getCharacterName(), c.IP.String())
			if c.is_mod {
				ret += " (mod)"
			}
		} else {
			ret += fmt.Sprintf("\r\n%s", c.getCharacterName())
		}
	}

	cl.sendServerMessageOOC(ret)
	writeClientLog(cl, "Used /getarea")
}

func cmdGetAllAreas(cl *Client) {
	var ret string
	for i := range config.Arealist {
		aptr := &config.Arealist[i]
		ret += fmt.Sprintf("\r\n=== Area %d: %s ===", aptr.Areaid, aptr.Name)
		for _, c := range aptr.sortedClientsByName() {
			if cl.is_mod {
				ret += fmt.Sprintf("\r\n%s; IP: %s", c.getCharacterName(), c.IP.String())
				if c.is_mod {
					ret += " (mod)"
				}
			} else {
				ret += fmt.Sprintf("\r\n%s", c.getCharacterName())
			}
		}
	}
	cl.sendServerMessageOOC(ret)
	writeClientLog(cl, "Used /getareas")
}

func cmdBackground(cl *Client, args []string) {
	if len(args) == 0 {
		cl.sendServerMessageOOC("The current background is " + cl.area.Background)
	} else if cl.area.bglock == true {
		if cl.is_mod == true {
			if err := cl.area.changeBackground(args[0]); err != nil {
				cl.sendServerMessageOOC("Background not found.")
			} else {
				cl.area.sendServerMessageOOC(fmt.Sprintf("%s changed background to %s.",
					cl.getCharacterName(), args[0]))
				writeClientLog(cl, "moderator changed locked background to "+args[0])
			}
		} else {
			cl.sendServerMessageOOC("A moderator has locked the background")
		}
	} else {
		if err := cl.area.changeBackground(args[0]); err != nil {
			cl.sendServerMessageOOC("Background not found.")
		} else {
			cl.area.sendServerMessageOOC(fmt.Sprintf("%s changed background to %s.",
				cl.getCharacterName(), args[0]))
			writeClientLog(cl, "changed background to "+args[0])
		}
	}
}

func cmdBgLock(cl *Client) {
	if cl.is_mod == true {
		if cl.area.bglock {
			cl.area.bglock = false
			cl.area.sendServerMessageOOC("Background unlocked.")
			writeClientLog(cl, "Unlocked the background.")
		} else {
			cl.area.bglock = true
			cl.area.sendServerMessageOOC("Background locked.")
			writeClientLog(cl, "Locked the background.")
		}
	} else {
		cl.sendServerMessageOOC("You do not have permission to use that")
	}
}

func cmdLogin(cl *Client, args []string) {
	if len(args) != 1 {
		cl.sendServerMessageOOC("Invalid arguments.")
		return
	}

	if args[0] == config.Modpass {
		cl.is_mod = true
		cl.sendServerMessageOOC("Logged in as a moderator.")
		writeClientLog(cl, "Logged in as a moderator.")
	} else {
		cl.sendServerMessageOOC("Invalid password.")
		writeClientLog(cl, "Tried to use mod login")
	}
}

func cmdIpList(cl *Client) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("You cannot use that command.")
		return
	}
	ret := fmt.Sprintf("\r\n=== Clients ===")
	for _, c := range client_list.sortedClientsByIP() {
		ret += fmt.Sprintf("\r\nIP: %s; %s in %s",
			c.IP.String(), c.getCharacterName(), c.getAreaName())
		if c.is_mod {
			ret += " (mod)"
		}
	}
	cl.sendServerMessageOOC(ret)
	writeClientLog(cl, "Used /iplist")
}

func cmdMute(cl *Client, target string) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}

	cnt := 0
	for _, v := range client_list.findAllTargets(cl, target) {
		if !v.muted {
			v.muted = true
			writeClientLog(cl, " muted"+v.IP.String())
			v.sendServerMessageOOC("You have been muted by a moderator.")
			cnt++
		}
	}

	if cnt == 0 {
		cl.sendServerMessageOOC("No unmuted targets found.")
	} else {
		cl.sendServerMessageOOC(fmt.Sprintf("Muted %d client(s).", cnt))
	}
}

func cmdUnmute(cl *Client, target string) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}

	cnt := 0
	for _, v := range client_list.findAllTargets(cl, target) {
		if v.muted {
			v.muted = false
			writeClientLog(cl, " unmuted"+v.IP.String())
			v.sendServerMessageOOC("You have been unmuted by a moderator.")
			cnt++
		}
	}

	if cnt == 0 {
		cl.sendServerMessageOOC("No muted targets found.")
	} else {
		cl.sendServerMessageOOC(fmt.Sprintf("Unmuted %d client(s).", cnt))
		writeClientLog(cl, fmt.Sprintf("Unmuted %d client(s).", cnt))
	}
}

func cmdKick(cl *Client, target string) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}

	cnt := 0
	for _, v := range client_list.findAllTargets(cl, target) {
		v.disconnect()
		writeClientLog(cl, "kicked "+v.getCharacterName()+"@"+v.getAreaName())
		cnt++
	}

	if cnt == 0 {
		cl.sendServerMessageOOC("No targets found.")
	} else {
		cl.sendServerMessageOOC(fmt.Sprintf("Kicked %d client(s).", cnt))
		writeClientLog(cl, fmt.Sprintf("kicked %d client(s)", cnt))
	}
}

func cmdBan(cl *Client, args []string) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}

	targetip := ""
	reason := "N/A"

	if len(args) >= 1 {
		targetip = args[0]
	} else {
		cl.sendServerMessageOOC("Argument must be an IP.")
		return
	}

	if len(args) > 1 {
		reason = strings.Join(args[1:], " ")
	}

	ban_clients := client_list.findTargetsByIP(cl, targetip)

	if len(ban_clients) == 0 {
		cl.sendServerMessageOOC("No targets found.")
	} else {
		ban_list.addBan(ban_clients[0], reason)
		cl.sendServerMessageOOC(fmt.Sprintf("Banned IP %s.", targetip))
		writeClientLog(cl, fmt.Sprintf("Banned IP %s. Reason: %s.",
			targetip, reason))
		for _, v := range ban_clients {
			v.disconnect()
		}
	}
}

func cmdReloadBans(cl *Client) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}

	if err := ban_list.loadBanlist(); err != nil {
		cl.sendServerMessageOOC(err.Error())
	} else {
		writeClientLog(cl, "Reloaded bans")
		cl.sendServerMessageOOC("Banlist reloaded.")
	}
}

func cmdSwitch(cl *Client, name string) {
	oldchar := cl.getCharacterName()
	if charid, err := getCIDfromName(name); err != nil {
		cl.sendServerMessageOOC(err.Error())
	} else {
		if err := cl.changeCharacterID(charid); err != nil {
			cl.sendServerMessageOOC(err.Error())
		} else {
			cl.sendServerMessageOOC("Successfully changed character.")
			writeClientLog(cl, fmt.Sprintf("Changed character from %s to %s.",
				oldchar, cl.getCharacterName()))
		}
	}
}

func cmdCharselect(cl *Client) {
	cl.charSelect()
}

func cmdRandomChar(cl *Client) {
	if cid, err := cl.area.randomFreeCharacterID(); err != nil {
		cl.sendServerMessageOOC(err.Error())
	} else {
		if err := cl.changeCharacterID(cid); err != nil {
			cl.sendServerMessageOOC(err.Error())
		} else {
			cl.sendServerMessageOOC(fmt.Sprintf("Randomly chose %s.",
				cl.getCharacterName()))
		}
	}
}

func cmdPM(cl *Client, target string) {
	var targets []*Client
	var name string
	var message string

	split_msg := strings.SplitN(target, " ", 2)
	if len(split_msg) != 2 {
		cl.sendServerMessageOOC("Invalid PM format.")
		return
	}

	charname, msg, err := msgStartsWithChar(target)
	if err == nil {
		if len(msg) == 0 {
			cl.sendServerMessageOOC("Message is empty.")
			return
		}

		name = charname
		message = msg
		if tgt := client_list.findTargetByChar(cl, name); tgt != nil {
			targets = append(targets, tgt)
		}
	} else {
		message = split_msg[1]
		name = split_msg[0]
		targets = client_list.findAllTargets(cl, name)
	}

	if len(targets) == 0 {
		cl.sendServerMessageOOC(fmt.Sprintf("Could not find %s.", name))
	}

	for _, v := range targets {
		v.sendServerMessageOOC(fmt.Sprintf("PM %s to You: %s", cl.oocname, message))
		cl.sendServerMessageOOC(fmt.Sprintf("PM You to %s: %s", name, message))
		writeClientLog(cl, fmt.Sprintf("Sent a PM to %s/%s in %s", v.getCharacterName(), v.oocname, v.getAreaName()))
	}
}

func cmdPos(cl *Client, target string) {
	if len(target) == 0 {
		cl.resetPos()
		cl.sendServerMessageOOC("Position reset.")
	} else {
		if err := cl.changePos(target); err != nil {
			cl.sendServerMessageOOC(err.Error())
		}
	}
}

func cmdGlobalMessage(cl *Client, message string) {
	if len(message) == 0 {
		cl.sendServerMessageOOC("Message is empty.")
	} else {
		client_list.sendAllRawIf(fmt.Sprintf(
			"CT#$GLOBAL[%v][%s]#%s#%", cl.area.Areaid, cl.getCharacterName(), message),
			func(c *Client) bool {
				return c.global
			})
		writeClientLog(cl, "[GLOBAL]"+message)
	}
}

func cmdGlobalToggle(cl *Client) {
	if cl.global {
		cl.global = false
		cl.sendServerMessageOOC("Global toggled off.")
	} else {
		cl.global = true
		cl.sendServerMessageOOC("Global toggled on.")
	}
}

func cmdNeed(cl *Client, message string) {
	if len(message) == 0 {
		cl.sendServerMessageOOC("Message is empty.")
	} else {
		client_list.sendAllRawIf(fmt.Sprintf(
			"CT#"+config.Reservedname+"#\r\n=======ADVERT=======\r\n"+cl.getCharacterName()+" in "+cl.getAreaName()+" needs "+message+"\r\n"+"===================#%"),
			func(c *Client) bool {
				return c.advert
			})
		writeClientLog(cl, "[NEED]"+message)
	}
}

func cmdAdvertToggle(cl *Client) {
	if cl.advert {
		cl.advert = false
		cl.sendServerMessageOOC("Adverts toggled off.")
	} else {
		cl.advert = true
		cl.sendServerMessageOOC("Adverts toggled on.")
	}
}

func cmdModAnnounce(cl *Client, message string) {
	if cl.is_mod != true {
		cl.sendServerMessageOOC("You do not have permission to use that")
	} else if len(message) == 0 {
		cl.sendServerMessageOOC("Message is empty.")
	} else {
		client_list.sendAllRaw("CT#$HOST#\r\n====ANNOUNCEMENT====\r\n------------------------------------\r\n" + message + "\r\n------------------------------------\r\n===================#%")
		writeClientLog(cl, "used Mod Announcement: "+message)
	}
}

func cmdMOTD(cl *Client, message string) {
	if len(message) == 0 {
		cl.sendServerMessageOOC("\r\n========MOTD========\r\n" + config.MOTD + "\r\n===================")
	} else if cl.is_mod {
		config.MOTD = message
		writeClientLog(cl, "changed the MOTD.")
	}
}

func cmdRoll(cl *Client, max string) {
	if len(max) == 0 {
		roll := randomInt(1, 6)
		cl.area.sendServerMessageOOC(cl.getCharacterName() + " rolled " + strconv.Itoa(roll) + " out of 6")
		writeClientLog(cl, "used roll")
	} else {
		maxroll, err := strconv.Atoi(max)
		if err != nil {
			cl.sendServerMessageOOC("The roll must be a number.")
			return
		}
		if maxroll > 1 && maxroll <= 9999 {
			roll := randomInt(1, maxroll)
			cl.area.sendServerMessageOOC(cl.getCharacterName() + " rolled " + strconv.Itoa(roll) + " out of " + max)
			writeClientLog(cl, "used roll")
		} else {
			cl.sendServerMessageOOC("The roll must be between 2 and 999.")
		}
	}
}

func cmdHelp(cl *Client) {
	cl.sendServerMessageOOC("A list of commands can be found here: https://github.com/AttorneyOnlineVidya/tsuserver/blob/master/README.md")
}

func cmdStatus(cl *Client, target string) {
	if len(target) == 0 {
		cl.sendServerMessageOOC("The area is currently set to " + cl.area.getAreaStatus())
	} else {
		switch strings.ToLower(target) {
		case "idle":
			cl.area.setAreaStatus(cl, "IDLE")
		case "buildingopen":
			cl.area.setAreaStatus(cl, "BUILDING-OPEN")
		case "buildingfull":
			cl.area.setAreaStatus(cl, "BUILDING-FULL")
		case "casingopen":
			cl.area.setAreaStatus(cl, "CASING-OPEN")
		case "casingfull":
			cl.area.setAreaStatus(cl, "CASING-FULL")
		case "recess":
			cl.area.setAreaStatus(cl, "RECESS")
		default:
			cl.sendServerMessageOOC("Couldn't recognize status. Try: idle, buildingopen, buildingfull, casingopen, casingfull, recess")
		}
	}
}

func cmdLocalMod(cl *Client, message string) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("You do not have permission to use that")
	} else if len(message) == 0 {
		cl.sendServerMessageOOC("Message is empty")
	} else {
		cl.area.sendRawMessage("CT#$MOD[" + cl.getCharacterName() + "]#" + message + "#%")
		writeClientLog(cl, "[LOCMOD]"+message)
	}
}

func cmdGlobalMod(cl *Client, message string) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("You do not have permission to use that")
	} else if len(message) == 0 {
		cl.sendServerMessageOOC("Message is empty")
	} else {
		fullmessage := fmt.Sprintf("CT#$GLOBAL[M][%v][%s]#%s#%", cl.area.Areaid, cl.getCharacterName(), message)
		client_list.sendAllRaw(fullmessage)
		writeClientLog(cl, "[GLOMOD]"+message)
	}
}

func cmdSetDoc(cl *Client, URL string) {
	if len(URL) == 0 {
		cl.sendServerMessageOOC("Message is empty.")
	} else {
		cl.area.setDoc(URL)
		cl.area.sendServerMessageOOC(cl.getCharacterName() + " changed the doc URL.")
		writeClientLog(cl, "changed the doc URL:"+URL)
	}
}

func cmdGetDoc(cl *Client) {
	cl.sendServerMessageOOC("Doc: " + cl.area.getDoc())
	writeClientLog(cl, "[DOC]HDID:"+cl.HDID)
}

func cmdNewPoll(cl *Client, target string) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("You do not have permission to use that.")
		return
	}

	if len(target) == 0 {
		cl.sendServerMessageOOC("Must specify a name.")
		return
	}

	if err := poll_list.newPoll(target); err != nil {
		cl.sendServerMessageOOC(err.Error())
	} else {
		client_list.sendAllRaw(fmt.Sprintf(
			"CT#%s#\r\n========POLL========\r\nA new poll called '%s' has been created. Use /vote to participate.\r\n===================#%%",
			config.Reservedname, target))
		writeClientLog(cl, "Created poll "+target)
		cl.sendServerMessageOOC("Poll created.")
	}
}

func cmdPollResults(cl *Client, target string) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("You do not have permission to use that.")
		return
	}

	if msg, err := poll_list.getPollResults(target); err != nil {
		cl.sendServerMessageOOC(err.Error())
	} else {
		writeClientLog(cl, "Checked poll results for "+target)
		cl.sendServerMessageOOC(msg)
	}
}

func cmdClosePoll(cl *Client, target string) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("You do not have permission to use that.")
		return
	}

	if err := poll_list.closePoll(target); err != nil {
		cl.sendServerMessageOOC(err.Error())
	} else {
		writeClientLog(cl, "Closed poll "+target)
		client_list.sendAllRaw(fmt.Sprintf(
			"CT#%s#\r\n========POLL========\r\nThe poll '%s' has been closed. Thank you for your votes.\r\n===================#%%",
			config.Reservedname, target))
		cl.sendServerMessageOOC("Poll closed.")
	}
}

func cmdPolls(cl *Client) {
	var ret string
	polls := poll_list.getPollList()
	if len(polls) == 0 {
		ret = "There are currently no available polls."
	} else {
		ret = "Currently available polls: "
		ret += strings.Join(polls, ", ")
		ret += ". Use /vote to cast your vote."
	}
	cl.sendServerMessageOOC(ret)
}

func cmdVote(cl *Client, args []string) {
	length := len(args)

	if length < 2 {
		cl.sendServerMessageOOC("Insufficient arguments. Usage: /vote [poll name] [yes/no].")
		return
	}

	poll_name := strings.Join(args[0:length-1], " ")
	vote_option := false

	if args[length-1] == "yes" {
		vote_option = true
	} else if args[length-1] == "no" {
		vote_option = false
	} else {
		cl.sendServerMessageOOC("Invalid vote, use yes or no.")
		return
	}

	if err := poll_list.vote(cl, poll_name, vote_option); err != nil {
		cl.sendServerMessageOOC(err.Error())
	} else {
		writeClientLog(cl, "Voted in poll "+poll_name)
		cl.sendServerMessageOOC("Vote cast / updated.")
	}
}

func cmdReloadPolls(cl *Client) {
	if !cl.is_mod {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}

	if err := poll_list.loadPolls(); err != nil {
		cl.sendServerMessageOOC(err.Error())
	} else {
		writeClientLog(cl, "Reloaded polls")
		cl.sendServerMessageOOC("Polls reloaded.")
	}
}
