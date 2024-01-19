## ewSysMgr

The ewSysMgr application exists to perform monitoring and alerting capabilites on k3s hosts. 
The binary runs on a cron job every 5 minutes, checking CPU, Memory, and Disk utilization
against the hard-coded threshold of 80%. If any of these values exceed these limits, the
system will fire a text message to my private phone number alerting me of which system and 
values tripped the circut. This will continue until the issue is resolved at 5 minute intervals.

This would easily be extensible in the future depending on growth/demands. 
