package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

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

func processRunCommand(command, room, sender string, cli *gomatrix.Client) {
	output, err := runCommand(command)
	if err != nil {
		cli.SendNotice(room, "Unable to run command: "+command)
	}
	cli.SendNotice(room, os.Getenv("NN_SERVER")+": "+command)
	cli.SendText(room, output)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Starting user", os.Getenv("NN_MATRIX_USER"), "on", os.Getenv("NN_SERVER"))
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
		if body == "!servers" {
			cli.SendText(ev.RoomID, "Running: "+os.Getenv("NN_MATRIX_USER")+" on "+os.Getenv("NN_SERVER"))
		}
		if body[0:2] == "! " {
			fmt.Println("["+ev.Sender+"]:", body[2:len(body)])
			body = body[2:len(body)]

		} else if body[0] == '!' && strings.Split(body, " ")[0][1:len(strings.Split(body, " ")[0])] == os.Getenv("NN_SERVER") {
			//  Optionally run on specific server. For example `!doom <command>`
			var splits = strings.Split(body, " ")
			fmt.Println("["+ev.Sender+"]:", strings.Join(splits[1:len(splits)], " "))
			body = strings.Join(splits[1:len(splits)], " ")
		} else {
			// Not for command execution
			return
		}
		processRunCommand(body, ev.RoomID, ev.Sender, cli)
	})

	if err := cli.Sync(); err != nil {
		fmt.Println("Sync() returned ", err)
	}
}
