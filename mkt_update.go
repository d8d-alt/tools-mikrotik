package main

import (
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
	serverName  = flag.String("ip", "", "Mikrotik IP")
	port        = flag.String("port", "", "Port")
	userName    = flag.String("user", "", "User Name")
	passWord    = flag.String("pass", "", "Password")
	update      = flag.Bool("update", false, "Update")
	maxAttempts = flag.Int("attemtps", 180, "max attemtps")
)

type VersMkt struct {
	currentPacketVers  string
	latestPacketVers   string
	currFirmwareVers   string
	latestFirmwareVers string
}

type SshCred struct {
	addressZ string
	client   *ssh.Client
	versMkt  VersMkt
}

func (c *SshCred) sshClient() (err error) {

	config := &ssh.ClientConfig{
		User: *userName,
		Auth: []ssh.AuthMethod{
			ssh.Password(*passWord),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	c.addressZ = net.JoinHostPort(*serverName, *port)

	for i := 1; i <= *maxAttempts; i++ {
		c.client, err = ssh.Dial("tcp", c.addressZ, config)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
		fmt.Printf("attemts to connect %d from max attempts %d\n", i, *maxAttempts)
	}

	return err
}

func (c *SshCred) sshExec(comm string) (execOut []byte, err error) {

	var session *ssh.Session
	for i := 1; i <= *maxAttempts; i++ {
		if err = c.sshClient(); err != nil {
			fmt.Println("There is an error with client refresh ... " + err.Error())
			return nil, err
		}

		session, err = c.client.NewSession()
		if err == nil {
			/*
				On "/system/routerboard/upgrade" needs new func to be created (becasue of mikrotik reboot) for new client, otherwise c.client must be reused.
				In the present moment on each command exec new client + new ssh session is creating. Maybe later I'll do some improvement, but it will work in this way fro now.
			*/
			defer c.client.Close()
			defer session.Close()

			execOut, err = session.CombinedOutput(comm)
			if err != nil {
				fmt.Println("There is an error with command exec (CombinedOutput) ... " + err.Error())
				c.client.Close()
				return nil, err
			}

			break
		}
		time.Sleep(1 * time.Second)
		if i == *maxAttempts {
			fmt.Printf("Reached max attemtps %d\n", *maxAttempts)
			return nil, err
		}
	}

	return execOut, nil
}

func (c *SshCred) valueAfter(line, key string) (string, bool) {
	value, ok := strings.CutPrefix(strings.TrimSpace(line), key)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(value), true
}

func (c *SshCred) getPacketsVersInfo() (needsUpdate bool, err error) {

	pVers, err := c.sshExec("/system/package/update/check-for-update")
	if err != nil {
		fmt.Println("There is an error to get Packets vers... " + err.Error())
		return false, err
	}

	for line := range strings.Lines(string(pVers)) {
		if value, ok := c.valueAfter(line, "installed-version:"); ok {
			c.versMkt.currentPacketVers = value
		}
		if value, ok := c.valueAfter(line, "latest-version:"); ok {
			c.versMkt.latestPacketVers = value
		}
	}

	return c.versMkt.currentPacketVers != c.versMkt.latestPacketVers, nil
}

func (c *SshCred) getFirmwareVersInfo() (needsUpdate bool, err error) {

	fVers, err := c.sshExec("/system routerboard print")
	if err != nil {
		fmt.Println("There is an error to get Firmware vers ... " + err.Error())
		return false, err
	}

	for line := range strings.Lines(string(fVers)) {
		if value, ok := c.valueAfter(line, "current-firmware:"); ok {
			c.versMkt.currFirmwareVers = value
		}
		if value, ok := c.valueAfter(line, "upgrade-firmware:"); ok {
			c.versMkt.latestFirmwareVers = value
		}
	}

	return c.versMkt.currFirmwareVers != c.versMkt.latestFirmwareVers, nil
}

func (c *SshCred) performUpdate(update *bool) error {
	// packets update part
	pVer, err := c.getPacketsVersInfo()
	if err != nil {
		fmt.Println("Cannot perform packet versions check ... " + err.Error())
		return err
	}

	if !pVer {
		fmt.Println("There is no new packets/firmware version for update")
		return nil
	} else {
		fmt.Printf("it needs to update packets from %s to %s : %v\n", c.versMkt.currentPacketVers, c.versMkt.latestPacketVers, pVer)

		if *update {
			if _, err = c.sshExec("/system/package/update/install"); err != nil {
				fmt.Println("Cannot perform package install from performUpdate func... " + err.Error())
				return err
			}

		}
	}

	// firmware update part

	fVer, err := c.getFirmwareVersInfo()
	if err != nil {
		fmt.Println("Cannot perform firmware versions check ... " + err.Error())
		return err
	}

	fmt.Printf("it needs to update firmware from %s to %s : %v\n :", c.versMkt.currFirmwareVers, c.versMkt.latestFirmwareVers, fVer)
	if fVer && *update {
		// sleep for a while till ping stop working on mikrotik reboot from packets update
		time.Sleep(5 * time.Second)
		if _, err = c.sshExec("/system/routerboard/upgrade"); err != nil {
			fmt.Println("Cannot perform routerboard upgrade from performUpdate func... " + err.Error())
			return err
		}
		if _, err = c.sshExec("/system reboot"); err != nil {
			fmt.Println("Cannot perform routerboard reboot after ROS upgrade from performUpdate func... " + err.Error())
			return err
		}
	}

	return nil
}

func main() {
	flag.Parse()
	if *serverName == "" || *port == "" || *userName == "" || *passWord == "" {
		log.Fatalf("usage: %s -ip=<ip> -port=<port> -user=<user> -pass=<pass> [-update=true]\n", filepath.Base(os.Args[0]))
	}

	var c SshCred

	err := c.performUpdate(update)
	if err != nil {
		fmt.Println("Cannot perform Update because of... " + err.Error())

	}
}
