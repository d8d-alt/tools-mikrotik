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

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	serverName  = flag.String("ip", "", "Mikrotik IP")
	port        = flag.String("port", "", "Port")
	userName    = flag.String("user", "", "User Name")
	passWord    = flag.String("pass", "", "Password")
	maxAttempts = flag.Int("attemtps", 60, "max attemtps")
)

type SshCred struct {
	addressZ string
	client   *ssh.Client
	hostName string
	fileTo   string
	fileBckp string
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
		} else {
			if strings.Contains(string(err.Error()), "unable to authenticate") {
				fmt.Println("it seems there is an issue with credentials:")
				return err
			}
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

func (c *SshCred) mktGetName() (mktName string, err error) {
	gName, err := c.sshExec("/system/identity/export compact")
	if err != nil {
		fmt.Printf("Failed to execute cmd fot Output... " + err.Error())
		return "", err
	}

	getLetter := string(gName)
	var namePrepare []string
	for _, v := range getLetter {
		namePrepare = append(namePrepare, string(v))
	}

	x := strings.TrimSpace(strings.Join(namePrepare, ""))
	if x == "" {
		fmt.Printf("Failed to trim name... " + err.Error())
		return "", err
	}
	_, mktName, _ = strings.Cut(x, "name=")
	return mktName, nil

}

func (c *SshCred) dateTime() string {
	return time.Now().Format("2006-01-02_150405")
}

func (c *SshCred) backupConf() error {

	c.fileTo = c.hostName + "_" + c.dateTime()

	_, err := c.sshExec("/system/backup/save name=" + c.fileTo)
	if err != nil {
		fmt.Printf("Failed to execute cmd fot Output... " + err.Error())
		return err
	}

	return nil
}

func (c *SshCred) exportConf() error {

	c.fileBckp = c.hostName + "_" + c.dateTime()

	_, err := c.sshExec("/export file=" + c.fileBckp + " compact")
	if err != nil {
		fmt.Printf("Failed to execute cmd fot Output... " + err.Error())
		return err
	}

	return nil
}

func (c *SshCred) bckCopy(sPath, dPath string) error {

	if err := c.sshClient(); err != nil {
		fmt.Printf("Failed to create new ssh client for sftp... " + err.Error())
		return err
	}

	sftp, err := sftp.NewClient(c.client)
	if err != nil {
		fmt.Printf("Failed to create new sftp client... " + err.Error())
		return err
	}
	defer sftp.Close()

	sFile, err := sftp.Open(sPath)
	if err != nil {
		fmt.Printf("Failed to open remote sftp file... " + err.Error())
		return err
	}

	dFile, err := os.Create(dPath)
	if err != nil {
		fmt.Printf("Failed to create local file... " + err.Error())
		return err
	}
	defer dFile.Close()

	sFile.WriteTo(dFile)

	err = sftp.Remove(sPath)
	if err != nil {
		fmt.Printf("Failed to remove remote sftp file... " + err.Error())
		return err
	}
	defer sFile.Close()
	return nil
}

func main() {
	flag.Parse()
	if len(os.Args) != 5 {
		log.Fatal("Error! Expected 4 arguments only! Exam: " + filepath.Base(os.Args[0]) + " -ip=192.168.253.1 -port=22 -user=username -pass=password")
	}

	if *serverName == "" || *port == "" || *userName == "" || *passWord == "" {
		log.Fatal("Error! All flags required: -ip, -port, -user, -pass")
	}

	var c SshCred
	var err error

	c.hostName, err = c.mktGetName()
	if err != nil {
		log.Fatalf("Cannot perform Update because of... " + err.Error())
	}

	if err = c.backupConf(); err != nil {
		log.Fatalf("Cannot save backup file localy in mikrotik... " + err.Error())
	}

	if err = c.exportConf(); err != nil {
		log.Fatalf("Cannot save backup file localy in mikrotik... " + err.Error())
	}

	c.fileBckp = c.fileBckp + ".rsc"
	c.fileTo = c.fileTo + ".backup"

	files := []string{c.fileBckp, c.fileTo}

	for _, file := range files {
		c.bckCopy(file, file)
		fmt.Printf("File : %s has been copied locally\n", file)
	}
}
