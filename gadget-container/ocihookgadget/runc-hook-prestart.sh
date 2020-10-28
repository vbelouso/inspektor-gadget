#!/bin/bash
read -r JSON
pidof gadgettracermanager > /dev/null || exit 0
echo $JSON | /opt/bin/ocihookgadget -socketfile /var/run/gadget/gadgettracermanager.socket -hook prestart >> /var/log/gadget.log 2>&1
exit 0
