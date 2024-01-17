package main

import (
	"tonysoft.com/gasp/gowatch/console"
	"tonysoft.com/gasp/gowatch/watch"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	err := console.Init()
	check(err)

	err = watch.Start()
	check(err)
}
