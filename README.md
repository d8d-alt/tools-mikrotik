Hi, 
... not sure on 100% this is the perfect backup / update for mikrotik over ssh/sftp but, they work for me... :)

... for mkt_update.go ( use -update=true to perform update, -update=false is just to check if there are new updates ):  <br>
<code>dim@RPi-171:~/bin $ ./mkt_update -ip=192.168.1.1 -port=22 -user=admin -pass=password -update=false
There is no new mikrotik firmware version for update... </code>


... for backup.go:  <br>
<code>dim@RPi-171:~/bin $ ./backup -ip=192.168.1.1 -port=22 -user=admin -pass=password
Configuration backup saved
File : <mkt_hostname>_2026-07-28_190204.backup.rsc has been copied locally
File : <mkt_hostname>_2026-07-28_190204.backup has been copied locally</code>
