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
			cl.area.Background = args[0]
			cl.area.sendRawMessage("BN#" + args[0] + "#%")
			cl.sendServerMessageOOC("changed background to " + args[0])
			writeClientLog(cl, "changed background to "+args[0])
		} else {
			cl.sendServerMessageOOC("A moderator has locked the background")
		}
	} else if stringInSlice(args[0], config.Backgroundlist) == true && cl.area.bglock == false {
		cl.area.Background = args[0]
		cl.area.sendRawMessage("BN#" + args[0] + "#%")
		cl.sendServerMessageOOC("changed background to " + args[0])
		writeClientLog(cl, "changed background to "+args[0])
	} else {
		cl.sendServerMessageOOC("That background cannot be found or is unavailable")
	}
}

func cmdBgLock(cl *Client, args []string) {
	if lock, err := strconv.ParseBool(args[0]); err == nil {
		if cl.is_mod == true {
			cl.area.bglock = lock
			cl.sendServerMessageOOC("Background lock is now " + args[0])
			writeClientLog(cl, "has set the background lock to "+args[0])
		} else {
			cl.sendServerMessageOOC("You do not have permission to use that")
		}
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
