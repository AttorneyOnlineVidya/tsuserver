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
			writeToLog(cl, "Changed area to "+cl.area.Name+".")
		}
	} else {
		cl.sendServerMessageOOC("Too many arguments.")
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
	} else {
		cl.sendServerMessageOOC("Invalid password.")
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
