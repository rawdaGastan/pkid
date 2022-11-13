package main

import (
	"github.com/rawdaGastan/pkid/internal"
)

func main() {

	const fileName = "pkid.db"
	internal.StartServer(fileName, "3000")
}
