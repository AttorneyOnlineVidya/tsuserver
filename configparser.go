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
	"log"
	"strconv"
	"strings"

	"github.com/burntsushi/toml"
)

var server_version string = "tsuserver0.1"

type Song struct {
	Name     string
	Duration int
}

type Config struct {
	Port           int
	Slots          int
	Charlist       []string
	Musiclist      []Song
	Arealist       []Area
	Backgroundlist []string
	Defaultarea    int
	Timeout        int
	Reservedname   string
	Modpass        string
	Guardpass      string
}

var config Config

func loadConfig() {
	// load configs
	if _, err := toml.DecodeFile("./config/config.toml", &config); err != nil {
		log.Fatal(err)
	}

	if _, err := toml.DecodeFile("./config/characters.toml", &config); err != nil {
		log.Fatal(err)
	}

	if _, err := toml.DecodeFile("./config/areas.toml", &config); err != nil {
		log.Fatal(err)
	}

	if _, err := toml.DecodeFile("./config/backgrounds.toml", &config); err != nil {
		log.Fatal(err)
	}

	if _, err := toml.DecodeFile("./config/secret.toml", &config); err != nil {
		log.Fatal(err)
	}

	var tmpconf struct{ Musiclist []string }

	if _, err := toml.DecodeFile("./config/musiclist.toml", &tmpconf); err != nil {
		log.Fatal(err)
	}

	for _, v := range tmpconf.Musiclist {
		spl := strings.Split(v, "#")
		name := spl[0]
		dur := -1
		if len(spl) == 2 {
			tmpdur, _ := strconv.Atoi(spl[1])
			dur = tmpdur
		}

		s := Song{Name: name, Duration: dur}
		config.Musiclist = append(config.Musiclist, s)
	}

	// set defaults
	for i := range config.Arealist {
		config.Arealist[i].initialize()
	}
}
