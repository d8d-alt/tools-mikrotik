package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	serverName = flag.String("ip", "", "Mikrotik IP")
	port       = flag.String("port", "", "Port")
	userName   = flag.String("user", "", "User Name")
	passWord   = flag.String("pass", "", "Password")
	update     = flag.Bool("update", false, "Update")
)

func conSSHserv() (session *ssh.Session) {

	flag.Parse()
	config := &ssh.ClientConfig{
		User: *userName,
		Auth: []ssh.AuthMethod{
			ssh.Password(*passWord),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshHst := *serverName + ":" + *port
	var err error
	client, err := ssh.Dial("tcp", sshHst, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	session, err = client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}

	return session
}

func stUpdate() {
	session := conSSHserv()
	if *update {
		err := session.Run("/system/package/update/install")
		if err != nil {
			log.Fatal("Failed to update router " + err.Error())
		}
	}

	defer session.Close()
	for {
		if strings.Contains(cHkOnline(), "ROSSSH") {
			if *update == true {
				updFirmware()
			} else {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func cHkOnline() string {
	flag.Parse()
	sshHst := *serverName + ":" + *port
	
	conn, err := net.Dial("tcp", sshHst)
	if err != nil {
		log.Fatal("Conn error from cHkOnline... " + err.Error())
	}
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Fatal("Cannot read new reader from cHkOnline... " + err.Error())
	}
	conn.Close()
	return status
}

func mktReboot() {
	session := conSSHserv()

	err := session.Run("/system reboot")
	if err != nil {
		log.Fatal("Failed to reboot  " + err.Error())
	}

	defer session.Close()
	os.Exit(0)
}

func updFirmware() {
	time.Sleep(15 * time.Second)
	session := conSSHserv()

	err := session.Run("/system/routerboard/upgrade")
	if err != nil {
		log.Fatal("Failed to update firmware " + err.Error())
	}

	defer session.Close()
	mktReboot()
}

func chkUpdate(update *bool) {
	session := conSSHserv()
	xout, err := session.Output("/system/package/update/check-for-updates")
	if err != nil {
		log.Fatal("Failed to execute cmd fot Output... " + err.Error())
	}

	defer session.Close()

	if strings.Contains(string(xout), "status: New version is available") {
		var shNew string
		var shCurr string
		for line := range strings.Lines(string(xout)) {
			if strings.Contains(line, "installed-version:") {
				shCurr = strings.Trim(strings.Join(strings.Split(line, "installed-version:"), ""), " \n\r")
			}

			if strings.Contains(line, "latest-version:") {
				shNew = strings.Trim(strings.Join(strings.Split(line, "latest-version:"), ""), " \n\r")
			}
		}

		if *update == false {
			fmt.Println("There is a new mikrotik firmware ver." + shNew + " and present installed is ver." + shCurr + ", please use -update=true to update it ... ")
		} else if *update == true {
			fmt.Println("Updating mikrotik firmware to ver." + shNew + " from present installed ver." + shCurr + "... ")
			stUpdate()
		}
	}

	if !strings.Contains(string(xout), "status: New version is available") {
		fmt.Println("There is no new mikrotik firmware version for update... ")
	}
}

func main() {
	flag.Parse()

	if *serverName == "" || *port == "" || *userName == "" || *passWord == "" {
		log.Fatalf("usage: %s -ip=<ip> -port=<port> -user=<user> -pass=<pass> [-update=true]\n", filepath.Base(os.Args[0]))
	}

	chkUpdate(update)
}
