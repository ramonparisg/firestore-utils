package main

import (
	"firestore-utils/internal/config"
)

func main() {
	ginController := config.GetControllersInstance()
	ginController.Start()
}
