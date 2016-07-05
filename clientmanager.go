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
	"net"
	"sync"
)

type Client struct {
	clientid uint64
	IP       net.IP
	conn     net.Conn
	charid   int
	area     *Area
	oocname  string
}

type ClientList struct {
	lock    sync.Mutex
	clients []*Client
}

// ================

func (cl *Client) changeAreaID(areaid int) error {
	// find the correct area pointer
	for i := range config.Arealist {
		v := &config.Arealist[i]
		if v.Areaid == areaid {
			// remove from old area if any
			if cl.area != nil {
				cl.area.removeClient(cl)
			}
			// add to new area
			v.addClient(cl)
			cl.area = v
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

// TODO fix senddone in general
func (cl *Client) sendDone() {
	/*
		CharsCheck - For each charid sends either a 0 if free or -1 if taken.
	*/
	charcheck := "CharsCheck"
	charcheck += "#0#%" // TODO fix charcheck

	cl.sendRawMessage(charcheck)
	cl.sendRawMessage("BN#gs4#%")
	cl.sendRawMessage("MM#1#%")
	//cl.sendRawMessage("OPPASS#676599#%") // TODO fix
	cl.sendRawMessage("DONE#%")
}

func (cl *Client) disconnect() {
	client_list.removeClient(cl)
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
