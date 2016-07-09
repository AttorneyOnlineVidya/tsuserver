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
	"net"
	"strconv"
	"sync"
)

type Client struct {
	clientid uint64
	IP       net.IP
	conn     net.Conn
	charid   int
	area     *Area
	oocname  string
	is_mod   bool
	muted    bool
	lock     sync.Mutex
}

type ClientList struct {
	lock    sync.Mutex
	clients []*Client
}

// ================

func (cl *Client) changeAreaID(areaid int) error {
	// check if the target area is the same
	if cl.area != nil && cl.area.Areaid == areaid {
		return errors.New("Target area is the same as the current one.")
	}
	// find the correct area pointer
	for i := range config.Arealist {
		v := &config.Arealist[i]
		if v.Areaid == areaid {
			cl.lock.Lock()
			defer cl.lock.Unlock()

			last_charid := cl.charid
			// change to a random character if taken
			if !v.isCharIDAvailable(cl.charid) && cl.charid != -1 {
				if id, err := v.randomFreeCharacterID(); err == nil {
					cl.charid = id
					cl.sendRawMessage("PV#" + strconv.FormatUint(cl.clientid, 10) +
						"#CID#" + strconv.Itoa(id) + "#%")
					cl.sendServerMessageOOC("Your character is taken, changing to a random one.")
				} else {
					return errors.New("Unable to switch, no free characters in target area.")
				}
			}
			// remove from old area if any
			if cl.area != nil {
				cl.area.removeClient(cl)
				cl.area.removeTakenCharacter(last_charid)
			}
			// add to new area
			v.addClient(cl)
			v.addTakenCharacter(cl.charid, cl)
			cl.area = v
			// send current penalties
			cl.sendRawMessage(fmt.Sprintf("HP#1#%d#%%", cl.area.hp_def))
			cl.sendRawMessage(fmt.Sprintf("HP#2#%d#%%", cl.area.hp_pro))
			return nil
		}
	}
	return errors.New("Target area does not exist.")
}

func (cl *Client) sendRawMessage(msg string) {
	cl.conn.Write([]byte(msg))
}

func (cl *Client) sendServerMessageOOC(msg string) {
	cl.sendRawMessage("CT#" + config.Reservedname + "#" + msg + "#%")
}

func (cl *Client) sendDone() {
	/*
		CharsCheck - For each charid sends either a 0 if free or -1 if taken.
	*/
	charcheck := "CharsCheck"
	for i := range config.Charlist {
		if cl.area.isCharIDAvailable(i) {
			charcheck += "#0"
		} else {
			charcheck += "#-1"
		}
	}
	charcheck += "#%"

	cl.sendRawMessage(charcheck)
	cl.sendRawMessage("BN#gs4#%")
	cl.sendRawMessage("MM#1#%")
	cl.sendRawMessage("OPPASS#" +
		encryptMessage(config.Guardpass, crypt_key) + "#%")
	cl.sendRawMessage("DONE#%")
}

func (cl *Client) changeCharacterID(id int) error {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	// check if available
	if cl.charid != id && !cl.area.isCharIDAvailable(id) {
		return errors.New("That character is unavailable.")
	}
	// add character to area
	cl.area.removeTakenCharacter(cl.charid)
	cl.charid = id
	cl.area.addTakenCharacter(id, cl)
	// send new character to user
	cl.sendRawMessage("PV#" + strconv.FormatUint(cl.clientid, 10) +
		"#CID#" + strconv.Itoa(cl.charid) + "#%")
	return nil
}

func (cl *Client) disconnect() {
	client_list.removeClient(cl)
	cl.area.removeTakenCharacter(cl.charid)
	cl.area.removeClient(cl)

	cl.conn.Close()
}

func (cl Client) getCharacterName() string {
	if isValidCharID(cl.charid) {
		return config.Charlist[cl.charid]
	}
	return ""
}

func (cl Client) getAreaName() string {
	if cl.area != nil {
		return cl.area.Name
	}
	return ""
}

func (cl Client) getPrintableAreaList() string {
	var ret string
	for _, a := range config.Arealist {
		cnt := a.getCharCount()
		ret += "\r\nArea " + strconv.Itoa(a.Areaid) + ": " +
			a.Name + " (" + strconv.Itoa(cnt) + " user"
		if cnt != 1 {
			ret += "s"
		}
		ret += ")"
		if cl.area.Areaid == a.Areaid {
			ret += " (*)"
		}
	}
	fmt.Println(ret)
	return ret
}

// ================

func (clist *ClientList) onlineCharacters() int {
	clist.lock.Lock()
	defer clist.lock.Unlock()

	count := 0
	for _, v := range clist.clients {
		if v.charid >= 0 {
			count += 1
		}
	}
	return count
}

func (clist *ClientList) addClient(c *Client) {
	clist.lock.Lock()
	defer clist.lock.Unlock()

	clist.clients = append(clist.clients, c)
}

func (clist *ClientList) removeClient(c *Client) {
	clist.lock.Lock()
	defer clist.lock.Unlock()

	for i, v := range clist.clients {
		if c == v {
			clist.clients = append(clist.clients[:i], clist.clients[i+1:]...)
			return
		}
	}
}

// returns the client who's using target character in the same area
func (clist *ClientList) findTargetByChar(cl *Client, target string) *Client {
	return cl.area.getClientByCharName(target)
}

func (clist *ClientList) findAllTargets(cl *Client, target string) []*Client {
	var ret []*Client

	if len(target) == 0 {
		return ret
	}

	if cl := clist.findTargetByChar(cl, target); cl != nil {
		ret = append(ret, cl)
	}

	return ret
}
