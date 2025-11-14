package main

import (
	"log"

	"github.com/cainlara/gozip/ui"
	"github.com/cainlara/gozip/util"
)

func main() {
	fileName, content, err := util.GetFileToExtract()
	if err != nil {
		log.Panic(err)
	}

	root := ui.BuildUI(fileName, content)

	if err := root.EnableMouse(false).Run(); err != nil {
		log.Panic(err)
	}
}
