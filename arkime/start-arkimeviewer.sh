#!/bin/bash

until curl -sS "http://$ES_HOST:$ES_PORT/_cluster/health?wait_for_status=yellow" > /dev/null 2>&1
do
    echo "Waiting for ES to start"
    sleep 3
done

echo "ES started..."

COUNT=1
ATTEMPTS=10
until [[ -f "/eve/eve.json" ]] || [[ $COUNT -gt $ATTEMPTS ]]
do
    echo "Waiting for Suricata to populate eve.json ( $(( COUNT++ ))/${ATTEMPTS} )"
    sleep 3
done
if [[ $COUNT -gt $ATTEMPTS ]]
then
    echo "Skipping Suricata eve.json..."
else
    echo "Suricata populated eve.json..."
fi

# set runtime environment variables
export ARKIME_ELASTICSEARCH="http://"$ES_HOST":"$ES_PORT
$ARKIMEDIR/db/db.pl $ARKIME_ELASTICSEARCH init --ifneeded
$ARKIMEDIR/bin/arkime_add_user.sh $ARKIME_ADMIN_USERNAME admin $ARKIME_ADMIN_PASSWORD --admin --webauth --email --remove --packetSearch --createOnly

echo "Starting Arkime capture in the background..."
cd $ARKIMEDIR
exec $ARKIMEDIR/bin/capture -c $ARKIMEDIR/etc/config.ini -R /data/ --monitor --skip 2>&1 &
echo " Default credentials"
echo " >> user: $ARKIME_ADMIN_USERNAME"
echo " >> password: $ARKIME_ADMIN_PASSWORD"

echo "Launch viewer..."
cd $ARKIMEDIR/viewer
$ARKIMEDIR/bin/node $ARKIMEDIR/viewer/viewer.js -c $ARKIMEDIR/etc/config.ini 2>&1
