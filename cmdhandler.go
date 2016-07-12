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
		cnt++
	}

	if cnt == 0 {
		cl.sendServerMessageOOC("No targets found.")
	} else {
		cl.sendServerMessageOOC(fmt.Sprintf("Kicked %d client(s).", cnt))
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
			"CT#$HOST#\r\n=======ADVERT=======\r\n"+cl.getCharacterName()+" in "+cl.getAreaName()+" needs "+message+"\r\n"+"===================#%"),
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
		client_list.sendAllAnnouncement(message)
		writeClientLog(cl, "used Mod Announcement: "+message)
	}
}

func cmdMOTD(cl *Client) {
	cl.sendServerMessageOOC("\r\n========MOTD========\r\n" + config.MOTD + "\r\n===================")
}
