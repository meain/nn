package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	"github.com/matrix-org/gomatrix"
)

func runCommand(command string) (string, error) {
	cmd := exec.Command(os.Getenv("NN_SHELL"), "-c", command)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return string(stdout), nil
}

func main() {
	fmt.Println("Starting nn...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Setting up user", os.Getenv("NN_MATRIX_USER"))
	cli, _ := gomatrix.NewClient(os.Getenv("NN_MATRIX_SERVER"), os.Getenv("NN_MATRIX_USER"), os.Getenv("NN_MATRIX_PASSWORD"))
	syncer := cli.Syncer.(*gomatrix.DefaultSyncer)

	syncer.OnEventType("m.room.message", func(ev *gomatrix.Event) {
		if ev.Sender == os.Getenv("NN_MATRIX_USER") {
			return
		}
		body, ok := ev.Body()
		if !ok {
			fmt.Println("Unable to parse body for request. Skipping.", ev.ID)
			return
		}
		fmt.Println("["+ev.Sender+"]:", body[2:len(body)])
		if body[0:2] == "! " {
			body = body[2:len(body)]
		} else {
			// Not for command execution
			return
		}
		output, err := runCommand(body)
		if err != nil {
			cli.SendNotice(ev.RoomID, "Unable to run command: "+body)
		}
		cli.SendText(ev.RoomID, output)
	})

	if err := cli.Sync(); err != nil {
		fmt.Println("Sync() returned ", err)
	}
}
