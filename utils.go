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
	"encoding/hex"
	"math"
	"math/rand"
	"strconv"
	"time"
)

var crypt_const1 uint16 = 53761
var crypt_const2 uint16 = 32618
var crypt_key uint16

var rng *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func calcKeys() {
	var tmp int
	tmp, _ = strconv.Atoi(decryptMessage([]byte("4"), 322))
	crypt_key = uint16(tmp)
}

func hexToBytes(hexstr string) []byte {
	out, _ := hex.DecodeString(hexstr)
	return out
}

func decryptMessage(enc_msg []byte, key uint16) string {
	// "crypt"
	var out string
	for _, val := range enc_msg {
		out += string(uint16(val) ^ (key >> 8))
		key = ((uint16(val) + key) * crypt_const1) + crypt_const2
	}
	return out
}

func encryptMessage(pt string, key uint16) string {
	// "crypt"
	var out string
	for _, chr := range pt {
		val := uint16(chr) ^ (key >> 8)
		out += strconv.FormatUint(uint64(val), 16)
		key = ((val + key) * crypt_const1) + crypt_const2
	}
	return out
}

func loadCharPages(perpage int) []string {
	var ret []string
	var str string = "CI#"

	for i, v := range config.Charlist {
		str += strconv.Itoa(i) + "#" + v + "&None&0&&&0&#"
		if math.Mod(float64(i), float64(perpage)) == float64(perpage-1) {
			str += "#%"
			ret = append(ret, str)
			str = "CI#"
		}
	}

	if len(str) > 3 {
		str += "#%"
		ret = append(ret, str)
	}

	return ret
}

func loadMusicPages(perpage int) []string {
	var ret []string
	var str string = "EM#"

	for i, v := range config.Musiclist {
		str += strconv.Itoa(i) + "#" + v.Name + "#"
		if math.Mod(float64(i), float64(perpage)) == float64(perpage-1) {
			str += "#%"
			ret = append(ret, str)
			str = "EM#"
		}
	}

	if len(str) > 3 {
		str += "#%"
		ret = append(ret, str)
	}

	return ret
}

func isValidCharID(id int) bool {
	return id >= 0 && id < len(config.Charlist)
}
