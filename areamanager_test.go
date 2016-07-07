package main

import (
	"testing"
	"time"
)

func BenchmarkTicker(b *testing.B) {
	var test_areas []Area
	for i := 0; i < 100; i++ {
		var a Area
		a.Areaid = i

		test_areas = append(test_areas, a)
	}

	ticker := time.NewTicker(time.Second * 1)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		go func() {
			<-ticker.C
			for i := 0; i < 100; i++ {
				test_areas[i].playMusic("testsong.mp3")
			}
		}()
		time.Sleep(10)
		ticker.Stop()
	}
}
