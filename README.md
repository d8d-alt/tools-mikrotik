Hi,

... for mkt_update.go ( use -update=true to perform update for both - packets and firmware ( with 2 reboots ) , -update=false is just to check if there are new updates or you can skip this option as it's set by default on false ):  <br><br>
<code>dim@RPi-171:~/bin $ ./mkt_update -ip=192.168.1.1 -port=22 -user=admin -pass=password -update=false
There is no new mikrotik firmware and packets version for update... </code>  <br><br>checked on my home mikrotik and it seems to work<br>
... it needs some code polishing to be performed
