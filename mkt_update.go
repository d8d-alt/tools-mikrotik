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
        serverName  = flag.String("ip", "", "Mikrotik IP")
        port        = flag.String("port", "", "Port")
        userName    = flag.String("user", "", "User Name")
        passWord    = flag.String("pass", "", "Password")
        update      = flag.Bool("update", false, "Update")
        maxAttempts = flag.Int("attemtps", 10, "max attemtps")
        // -- z vars
        addressZ    string
        dialTimeout = time.Duration(1 * time.Second)
)

func conSSHserv() (session *ssh.Session, err error) {

        config := &ssh.ClientConfig{
                User: *userName,
                Auth: []ssh.AuthMethod{
                        ssh.Password(*passWord),
                },
                HostKeyCallback: ssh.InsecureIgnoreHostKey(),
        }

        for i := 1; i <= *maxAttempts; i++ {
                if mConn, err := net.DialTimeout("tcp", addressZ, dialTimeout); err == nil {
                        mConn.Close()
                        break
                        // return nil, err
                }
                time.Sleep(1 * time.Second)
        }

        // var client *ssh.Client

        if client, err = ssh.Dial("tcp", addressZ, config); err != nil {
                fmt.Println("Failed to dial after configured maxAttempts checks: ", err.Error())
                return nil, err
        }

        if session, err = client.NewSession(); err != nil {
                fmt.Println("Failed to create session: ", err.Error())
                return nil, err
        }

        return session, err
}

func stUpdate() (err error) {

        session, err := conSSHserv()
        if err != nil {
                fmt.Println("There is an error with session creation... " + err.Error())
                return err
        }

        defer session.Close()

        if *update {
                if err := session.Run("/system/package/update/install"); err != nil {
                        fmt.Println("Failed to update router " + err.Error())
                        return err
                }
        }

        for {
                chkOnlineStatus, err := cHkOnline()
                if err != nil {
                        fmt.Println("Conn error from cHkOnline... " + err.Error())
                        return err
                }

                if strings.Contains(chkOnlineStatus, "ROSSSH") {
                        if *update == true {
                                if err := updFirmware(); err != nil {
                                        fmt.Println("Failed with updFirmware fuction... : " + err.Error())
                                        return err
                                }
                        } else {
                                break
                        }
                }
                time.Sleep(1 * time.Second)
        }

        return nil
}

func cHkOnline() (status string, err error) {
        time.Sleep(5 * time.Second)

        for i := 1; i <= *maxAttempts; i++ {
                fmt.Printf("Attempt to connect after reboot %d/%d: Checking %s...\n", i, *maxAttempts, addressZ)

                conn, err := net.DialTimeout("tcp", addressZ, dialTimeout)
                if err == nil {
                        status, err = bufio.NewReader(conn).ReadString('\n')
                        if err != nil {
                                fmt.Println("Cannot read new reader from cHkOnline..." + err.Error())
                                return "", err
                        }
                        conn.Close()
                        return status, nil
                }
                time.Sleep(1 * time.Second)
        }

        time.Sleep(1 * time.Second)

        return status, nil
}

func mktReboot() (err error) {
        session, err := conSSHserv()
        if err != nil {
                fmt.Println("There is an error with session creation... " + err.Error())
                return err
        }

        defer session.Close()

        if err = session.Run("/system reboot"); err != nil {
                fmt.Println("Failed to reboot  " + err.Error())
                return err
        }

        return nil

}

func updFirmware() (err error) {
        session, err := conSSHserv()
        if err != nil {
                fmt.Println("There is an error with session creation... " + err.Error())
                return err
        }

        defer session.Close()

        if err = session.Run("/system/routerboard/upgrade"); err != nil {
                fmt.Println("Failed to update firmware " + err.Error())
                return err
        }

        if err = mktReboot(); err != nil {
                log.Fatal("Cannot reboot mikrotik " + err.Error())
        } else {
                os.Exit(0)
        }

        return nil
}

func chkFirmware() (resp []byte, err error) {

        session, err := conSSHserv()
        if err != nil {
                fmt.Println("There is an error with session creation... " + err.Error())
                return nil, err
        }

        defer session.Close()

        firmwOut, err := session.Output("/system routerboard print")
        if err != nil {
                fmt.Println("Failed to execute cmd fot Output..." + err.Error())
                return nil, err
        }

        return firmwOut, nil
}

func chkPackets() (resp []byte, err error) {
        session, err := conSSHserv()
        if err != nil {
                fmt.Println("There is an error with session creation... " + err.Error())
                return nil, err
        }

        defer session.Close()

        packOut, err := session.Output("/system/package/update/check-for-updates")
        if err != nil {
                fmt.Println("Cannot check for updates..." + err.Error())
                return nil, err
        }

        return packOut, nil
}

func chkUpdate() (err error) {

        var (
                shFirmwareNew  string
                shFirmwareCurr string
                frmOut         []byte
        )

        if frmOut, err = chkFirmware(); err != nil {
                fmt.Println("Failed to get info for firmware... " + err.Error())
                return err
        }

        for lineZ := range strings.Lines(string(frmOut)) {
                if strings.Contains(lineZ, "current-firmware:") {
                        shFirmwareCurr = strings.Trim(strings.Join(strings.Split(lineZ, "current-firmware:"), ""), " \n\r")
                }

                if strings.Contains(lineZ, "upgrade-firmware:") {
                        shFirmwareNew = strings.Trim(strings.Join(strings.Split(lineZ, "upgrade-firmware:"), ""), " \n\r")
                }
        }

        if !strings.Contains(shFirmwareNew, shFirmwareCurr) {
                fmt.Printf("Current used firmware version %s needs to be updated to %s  \n", shFirmwareCurr, shFirmwareNew)

                if *update == true {
                        fmt.Println("Updating mikrotik to ver." + shFirmwareNew + " from current installed ver." + shFirmwareCurr + "... ")

                        err = updFirmware()
                        if err != nil {
                                log.Fatal("There is an error with firmware update... " + err.Error())
                        }
                }
        }

        var (
                shNew  string
                shCurr string
                xout   []byte
        )

        if xout, err = chkPackets(); err != nil {
                fmt.Println("Failed to get info for packets... " + err.Error())
                return err
        }

        if strings.Contains(string(xout), "status: New version is available") {

                for line := range strings.Lines(string(xout)) {
                        if strings.Contains(line, "installed-version:") {
                                shCurr = strings.Trim(strings.Join(strings.Split(line, "installed-version:"), ""), " \n\r")
                        }

                        if strings.Contains(line, "latest-version:") {
                                shNew = strings.Trim(strings.Join(strings.Split(line, "latest-version:"), ""), " \n\r")
                        }
                }

                if *update == false {
                        fmt.Println("There is a new mikrotik / packets ver." + shNew + " and current installed is ver." + shCurr + ", please use -update=true to update it ... ")
                } else if *update == true {
                        fmt.Println("Updating mikrotik to ver." + shNew + " from current installed ver." + shCurr + "... ")
                        stUpdate()
                }

        }

        if !strings.Contains(string(xout), "status: New version is available") {
                fmt.Println("There is no new mikrotik version for update... ")
        }

        return nil
}

func main() {
        flag.Parse()

        addressZ = net.JoinHostPort(*serverName, *port)

        if *serverName == "" || *port == "" || *userName == "" || *passWord == "" {
                log.Fatalf("usage: %s -ip=<ip> -port=<port> -user=<user> -pass=<pass> [-update=true]\n", filepath.Base(os.Args[0]))
        }

        if err := chkUpdate(); err != nil {
                log.Fatal("Failed to check for update " + err.Error())
        }
}
