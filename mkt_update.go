package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

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

}

func chkUpdate() {
	session := conSSHserv()
	xout, err := session.Output("/system/package/update/check-for-updates")
	if err != nil {
		log.Fatal("Failed to execute cmd fot Output... " + err.Error())
	}

	defer session.Close()

	if strings.Contains(string(xout), "status: New version is available") {
		var shCurr string
		var shNew string
		for line := range strings.Lines(string(xout)) {
			if strings.Contains(line, "installed-version:") {
				_, shCurr, _ = strings.Cut(line, "installed-version: ")
				shCurr = strings.Trim(shCurr, "\r\n ")
			}
			if strings.Contains(line, "latest-version:") {
				_, shNew, _ = strings.Cut(line, " latest-version: ")
				shNew = strings.Trim(shNew, "\r\n ")
			}
		}

		fmt.Println("There is a new mikrotik firmware ver." + shNew + " for update, going to update it from present installed ver." + shCurr + "... ")
		stUpdate()
	}

	if !strings.Contains(string(xout), "status: New version is available") {
		fmt.Println("There is no new mikrotik firmware version for update... ")
	}

}

func main() {

  	if len(os.Args) != 6 {
		log.Fatal("Error! Expected 5 arguments only! Exam:  ./mkt_update -ip=192.168.253.1 -port=22 -user=username -pass=password -update=true ")
	}

	var err error
	if serverName == nil {
		log.Fatal("ip not set: ", err)
	}
	if port == nil {
		log.Fatal("port not set: ", err)
	}
	if userName == nil {
		log.Fatal("user not set: ", err)
	}
	if passWord == nil {
		log.Fatal("pass not set: ", err)
	}
  if update == nil {
		log.Fatal("update not set: ", err)
	}



  

	chkUpdate()

}
