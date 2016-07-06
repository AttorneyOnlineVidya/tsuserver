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
	"sync"
)

type Area struct {
	Areaid  int
	Name    string
	clients []*Client
	lock    sync.Mutex
	hp_def  int
	hp_pro  int
}

func (a *Area) sendRawMessage(msg string) {
	client_list.lock.Lock()
	defer client_list.lock.Unlock()

	for _, v := range client_list.clients {
		if v.area == a {
			v.sendRawMessage(msg)
		}
	}
}

func (a *Area) getCharCount() int {
	a.lock.Lock()
	defer a.lock.Unlock()

	count := 0
	for _, c := range a.clients {
		if isValidCharID(c.charid) {
			count += 1
		}
	}

	return count
}

func (a *Area) addClient(c *Client) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.clients = append(a.clients, c)
}

func (a *Area) removeClient(c *Client) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for i, v := range a.clients {
		if c == v {
			a.clients = append(a.clients[:i], a.clients[i+1:]...)
			return
		}
	}
}

func (a *Area) setDefaults() {
	a.hp_def = 10
	a.hp_pro = 10
}

func (a *Area) setDefHP(hp int) error {
	if hp >= 0 && hp <= 10 {
		a.hp_def = hp
		return nil
	} else {
		return errors.New("Invalid HP value.")
	}
}

func (a *Area) setProHP(hp int) error {
	if hp >= 0 && hp <= 10 {
		a.hp_pro = hp
		return nil
	} else {
		return errors.New("Invalid HP value.")
	}
}
