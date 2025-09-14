#!/bin/sh
ssh-keygen -t ed25519 -f /root/ctf_toolbox/mover_ssh_key -N ''
cat /root/ctf_toolbox/mover_ssh_key
cat << EOF | tee /etc/cron.d/pcap_sync
*/1 * * * * root /usr/bin/rsync -av --inplace --ignore-existing --include='*.pcap' --exclude='*.incomplete' --exclude='*' -e "ssh -i /root/ctf_toolbox/mover_ssh_key" root@51.68.144.13:/opt/tcpdump/ /root/ctf_toolbox/_data/
EOF