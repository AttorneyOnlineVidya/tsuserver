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
	"math/rand"
	"strconv"
	"strings"
)

func cmdArea(cl *Client, args []string) {
	if len(args) == 0 {
		cl.sendServerMessageOOC(cl.getPrintableAreaList())
	} else if len(args) == 1 || len(args) == 2 {
		args = append(args, "")
		targetarea, err := strconv.Atoi(args[0])
		if err != nil {
			cl.sendServerMessageOOC("The argument must be a number.")
			return
		}
		if err := cl.changeAreaID(targetarea, args[1]); err != nil {
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
		if cl.isMod() {
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
		if !cl.isMod() && aptr.IsHidden {
			continue
		}
		ret += fmt.Sprintf("\r\n=== Area %d: %s ===", aptr.Areaid, aptr.Name)
		for _, c := range aptr.sortedClientsByName() {
			if cl.isMod() {
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
		if cl.isMod() == true {
			if err := cl.area.changeBackground(args[0], cl.isMod()); err != nil {
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
		if err := cl.area.changeBackground(args[0], cl.isMod()); err != nil {
			cl.sendServerMessageOOC("Background not found.")
		} else {
			cl.area.sendServerMessageOOC(fmt.Sprintf("%s changed background to %s.",
				cl.getCharacterName(), args[0]))
			writeClientLog(cl, "changed background to "+args[0])
		}
	}
}

func cmdBgLock(cl *Client) {
	if cl.isMod() == true {
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
		cl.setMod(true)
		cl.sendServerMessageOOC("Logged in as a moderator.")
		writeClientLog(cl, "Logged in as a moderator.")
	} else {
		cl.sendServerMessageOOC("Invalid password.")
		writeClientLog(cl, "Tried to use mod login")
	}
}

func cmdIpList(cl *Client) {
	if !cl.isMod() {
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
	if !cl.isMod() {
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
	if !cl.isMod() {
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

func cmdDJ(cl *Client, target string) {
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}

	cnt := 0
	for _, v := range client_list.findAllTargets(cl, target) {
		if !v.dj {
			v.dj = true
			writeClientLog(cl, " DJ'd "+v.IP.String())
			v.sendServerMessageOOC("You have been DJ'd by a moderator.")
			cnt++
		}
	}

	if cnt == 0 {
		cl.sendServerMessageOOC("No DJ targets found.")
	} else {
		cl.sendServerMessageOOC(fmt.Sprintf("DJ'd %d client(s).", cnt))
	}
}

func cmdUnDJ(cl *Client, target string) {
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}

	cnt := 0
	for _, v := range client_list.findAllTargets(cl, target) {
		if v.dj {
			v.dj = false
			writeClientLog(cl, " unDJ'd "+v.IP.String())
			v.sendServerMessageOOC("You have been unDJ'd by a moderator.")
			cnt++
		}
	}

	if cnt == 0 {
		cl.sendServerMessageOOC("No unDJ targets found.")
	} else {
		cl.sendServerMessageOOC(fmt.Sprintf("UnDJ'd %d client(s).", cnt))
	}
}

func cmdKick(cl *Client, target string) {
	if !cl.isMod() {
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
	if !cl.isMod() {
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
	if !cl.isMod() {
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

func cmdCharselect(cl *Client, target string) {
	if len(target) == 0 {
		cl.charSelect()
	} else if cl.isMod() {
		cnt := 0
		for _, v := range client_list.findAllTargets(cl, target) {
			v.charSelect()
			writeClientLog(cl, " charselected "+v.IP.String())
			v.sendServerMessageOOC("A moderator forced you into character selection.")
			cnt++
		}
		if cnt == 0 {
			cl.sendServerMessageOOC("No targets found.")
		} else {
			cl.sendServerMessageOOC(fmt.Sprintf("Forced %d client(s) into character selection.", cnt))
		}
	} else {
		cl.sendServerMessageOOC("Insufficient permissions.")
	}
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
	if cl.isMod() != true {
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
	} else if cl.isMod() {
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

func cmdCoinFlip(cl *Client) {
	switch rand.Intn(2) {
	case 0:
		cl.area.sendServerMessageOOC(cl.getCharacterName() + " flipped a coin and got heads.")
	case 1:
		cl.area.sendServerMessageOOC(cl.getCharacterName() + " flipped a coin and got tails.")
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
	if !cl.isMod() {
		cl.sendServerMessageOOC("You do not have permission to use that")
	} else if len(message) == 0 {
		cl.sendServerMessageOOC("Message is empty")
	} else {
		cl.area.sendRawMessage("CT#$MOD[" + cl.getCharacterName() + "]#" + message + "#%")
		writeClientLog(cl, "[LOCMOD]"+message)
	}
}

func cmdGlobalMod(cl *Client, message string) {
	if !cl.isMod() {
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
	if !cl.isMod() {
		cl.sendServerMessageOOC("You do not have permission to use that.")
		return
	}

	split_msg := strings.SplitN(target, " ", 2)
	if len(split_msg) != 2 {
		cl.sendServerMessageOOC("Invalid poll format.")
		return
	}
	description := split_msg[1]
	name := split_msg[0]

	if err := poll_list.newPoll(name, description); err != nil {
		cl.sendServerMessageOOC(err.Error())
	} else {
		client_list.sendAllRaw(fmt.Sprintf(
			"CT#%s#\r\n========POLL========\r\nA new poll called %s has been created. It's description is: %s \r\nUse /vote to participate.\r\n===================#%%",
			config.Reservedname, name, description))
		writeClientLog(cl, "Created poll "+name+" with the description: "+description)
		cl.sendServerMessageOOC("Poll created.")
	}
}

func cmdPollResults(cl *Client, target string) {
	if !cl.isMod() {
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

func cmdPollInfo(cl *Client, target string) {
	if len(target) == 0 {
		cl.sendServerMessageOOC("Must specify a poll title.")
		return
	}
	if desc, err := poll_list.getPollDescription(target); err != nil {
		cl.sendServerMessageOOC(err.Error())
	} else {
		cl.sendServerMessageOOC(desc)
	}

}

func cmdClosePoll(cl *Client, target string) {
	if !cl.isMod() {
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
		ret += ".\r\nUse /pollinfo poll to see the description. Use /vote to cast your vote."
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
	if !cl.isMod() {
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

func cmdJudgeLog(cl *Client) {
	var aptr = cl.getAreaPtr()
	var ret string
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}
	ret += fmt.Sprintf("=== Area %d: %s ===", aptr.Areaid, aptr.Name)
	for _, v := range aptr.HPLog {
		ret += fmt.Sprintf("\r\n%s", v)
	}

	cl.sendServerMessageOOC(ret)

}

func cmdModPlay(cl *Client, songname string) {
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}
	cl.area.playMusic(songname, cl.charid, -1)
	writeClientLog(cl, " changed music to "+songname)
}

func cmdReloadMusic(cl *Client) {
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}
	reloadMusicConfig()
	cl.sendServerMessageOOC("Musiclist reloaded, restart your client.")
	writeClientLog(cl, " reloaded the musiclist.")
}

func cmdLockArea(cl *Client, password string) {
	if len(password) == 0 {
		cl.sendServerMessageOOC("Must specify a password.")
		return
	}
	if !cl.isMod() {
		if !cl.area.canbelocked {
			cl.sendServerMessageOOC("You cannot lock this area")
			return
		}
	} else if cl.area.Areaid == config.Defaultarea {
		cl.sendServerMessageOOC("You cannot lock the default area.")
		return
	}
	cl.area.ispassworded = true
	cl.area.password = password
	cl.sendServerMessageOOC("The area is now locked with the password: " + password)
	cl.area.sendServerMessageOOC(cl.getCharacterName() + " locked the area.")
	writeClientLog(cl, " locked the area.")
}

func cmdUnlockArea(cl *Client) {
	if !cl.area.ispassworded {
		cl.sendServerMessageOOC("This area is not locked.")
		return
	}
	cl.area.ispassworded = false
	cl.sendServerMessageOOC("This area is now unlocked.")
	cl.area.sendServerMessageOOC(cl.getCharacterName() + " unlocked the area.")
	writeClientLog(cl, " unlocked the area.")
}

func cmdLockableArea(cl *Client) {
	if !cl.isMod() {
		cl.sendServerMessageOOC("You must be a moderator to use that command.")
		return
	}
	if cl.area.Areaid == config.Defaultarea {
		cl.sendServerMessageOOC("You cannot lock the default area.")
		return
	}
	if cl.area.canbelocked {
		cl.area.canbelocked = false
		cl.sendServerMessageOOC("This area is now not lockable.")
		writeClientLog(cl, " changed the area to not lockable.")
	} else {
		cl.area.canbelocked = true
		cl.sendServerMessageOOC("This area is now lockable.")
		writeClientLog(cl, " changed the area to lockable.")
	}
}

func cmdReloadCharlist(cl *Client) {
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}
	reloadCharList()
	cl.sendServerMessageOOC("Character list reloaded, restart your client.")
	writeClientLog(cl, " reloaded the charlist.")
}

func cmdReloadConfig(cl *Client) {
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}
	reloadConfig()
	cl.sendServerMessageOOC("Config reloaded.")
	writeClientLog(cl, " reloaded the config.")
}

func cmdReloadBackgrounds(cl *Client) {
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}
	reloadCharList()
	cl.sendServerMessageOOC("Background list reloaded.")
	writeClientLog(cl, " reloaded the bglist.")
}

func cmdReloadEvidence(cl *Client) {
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}
	reloadEvidence()
	cl.sendServerMessageOOC("Evidence list reloaded, restart your client")
	writeClientLog(cl, " reloaded the evidence list.")
}

func cmdMasterServerAdvertising(cl *Client) {
	advertising := isAdvertising()
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}
	if advertising {
		setAdvertising(false)
		cl.sendServerMessageOOC("Master server advertising has been stopped.")
		writeClientLog(cl, "Stopped advertising on the master server.")
	} else {
		setAdvertising(true)
		go msAdvertiser()
		cl.sendServerMessageOOC("Master server advertising has been started.")
		writeClientLog(cl, "started advertising on the master server.")
	}
}

func cmdGiveEvidence(cl *Client, evirequest string) {
	evinumber, err := strconv.Atoi(evirequest)
	if err != nil {
		cl.sendServerMessageOOC("The evidence must be a number.")
		return
	}
	if evinumber <= 0 || evinumber > (len(config.Evidencelist)+len(cust.Evidencelist)) {
		cl.sendServerMessageOOC("Could not find that evidence, please use a number between 1 and " + strconv.Itoa(len(config.Evidencelist)+len(cust.Evidencelist)))
		return
	}
	cl.sendRawMessage("MS#chat#normal#Hawk#tie#Evidence " + evirequest + "#jud#1#0#0#0#0#" + evirequest + "#0#0#3#%")
}

func cmdCreateEvidence(cl *Client, evistring string) {
	split_evi := strings.Split(evistring, "<num>")
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}

	if len(split_evi) != 4 {
		cl.sendServerMessageOOC("The evidence must be seperated with '#'")
		return
	}

	evi := Evidence{Name: split_evi[0], Desc: split_evi[1], Type: split_evi[2], Image: split_evi[3]}
	cust.AddEvidence(evi)
	newlength := len(cust.Evidencelist) + len(config.Evidencelist)
	cl.sendServerMessageOOC("Evidence created, evidence number is " + strconv.Itoa(newlength))
	writeClientLog(cl, "created a new piece of evidence")
}

func cmdClearCustomEvidence(cl *Client) {
	if !cl.isMod() {
		cl.sendServerMessageOOC("Invalid command.")
		return
	}

	cust.Evidencelist = nil
	cl.sendServerMessageOOC("Custom evidence list cleared.")
	writeClientLog(cl, "cleared the custom evidence list.")
}
