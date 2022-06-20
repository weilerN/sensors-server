package main

import (
	"SensorServer/internal/client"
	"fmt"
	"log"
)

func main() {
	conn := client.ConnectClient()

	client.MenuLoop()

	defer func() { //cleanup
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	fmt.Println("exit..")
}

//TODO
/*


 */
