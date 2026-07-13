package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	serverName = flag.String("ip", "", "Mikrotik IP")
	port       = flag.String("port", "", "Port")
	userName   = flag.String("user", "", "User Name")
	passWord   = flag.String("pass", "", "Password")
)

func clientSSHserv() (client *ssh.Client) {
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
	client, err = ssh.Dial("tcp", sshHst, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}

	return client

}

func conSSHserv() (session *ssh.Session) {
	session, err := clientSSHserv().NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}

	return session
}

func mktGetName() (mktName string) {
	session := conSSHserv()
	gName, err := session.Output("/system/identity/print")
	if err != nil {
		log.Fatal("Failed to execute cmd fot Output... " + err.Error())
	}

	defer session.Close()
	gNameZ := string(gName)
	_, gNameZ, _ = strings.Cut(gNameZ, "name:")
	gNameZ = strings.Trim(gNameZ, " \n\r")
	return gNameZ

}

func backupConf(datetime, name string) (fileTo string) {
	session := conSSHserv()
	fileTo = name + "_" + datetime

	bckpOut, err := session.Output("/system/backup/save name=" + fileTo)
	if err != nil {
		log.Fatal("Failed to execute cmd fot Output... " + err.Error())
	}

	defer session.Close()
	fmt.Printf("%v\n", string(bckpOut))
	return fileTo
}

func exportConf(datetime, name string) (fileTo string) {
	session := conSSHserv()
	fileTo = name + "_" + datetime + ".backup"

	xout, err := session.Output("/export file=" + fileTo + " compact")
	if err != nil {
		log.Fatal("Failed to execute cmd fot Output... " + err.Error())
	}

	defer session.Close()
	fmt.Printf("%v\n", string(xout))
	return fileTo
}

func bckCopy(sPath, dPath string) {
	client := clientSSHserv()

	sftp, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal("Failed to create new sftp client... " + err.Error())
	}
	defer sftp.Close()

	sFile, err := sftp.Open(sPath)
	if err != nil {
		log.Fatal("Failed to open remote sftp file... " + err.Error())
	}
	defer sFile.Close()

	dFile, err := os.Create(dPath)
	if err != nil {
		log.Fatal("Failed to create local file... " + err.Error())
	}
	defer dFile.Close()

	sFile.WriteTo(dFile)

	err = sftp.Remove(sPath)
	if err != nil {
		log.Fatal("Failed to remove remote sftp file... " + err.Error())
	}
	defer sFile.Close()
}

func main() {

	if len(os.Args) != 5 {
		log.Fatal("Error! Expected 4 arguments only! Exam:  ./backup -ip=192.168.253.1 -port=22 -user=username -pass=password")
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

	datetime := time.Now().Format("2006-01-02_150405")
	mktNm := mktGetName()

	var files []string
	expFile := exportConf(datetime, mktNm) + ".rsc"
	bckFile := backupConf(datetime, mktNm) + ".backup"

	files = append(files, expFile, bckFile)

	for _, file := range files {
		bckCopy(file, file)
		fmt.Printf("File : %s has been copied locally\n", file)
	}

}
