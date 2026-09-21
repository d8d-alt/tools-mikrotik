... for mkt_update.go ( use -update=true to perform update for both - packets and firmware ( with 2 reboots ) , -update=false is just to check if there are new updates or you can skip this option as it's set by default on false ):  <br><br>
<code>dim@RPi-171:~/bin $ ./mkt_update -ip=192.168.1.1 -port=22 -user=admin -pass=password -update=false
There is no new mikrotik firmware and packets version for update... </code>  <br><br>checked on my home mikrotik and it seems to work<br>
... it needs some code polishing to be performed <br><br><br>


... for backup.go:<br>
<code>dim@RPI-169:~/$ backup -ip=192.168.253.1 -port=22 -user=user -pass=password
File : mikrotik-hostname1_2026-09-21_232506.rsc has been copied locally
File : mikrotik-hostname2_2026-09-21_232504.backup has been copied locally</code>
