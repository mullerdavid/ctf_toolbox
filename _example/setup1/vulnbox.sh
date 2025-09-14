#!/bin/sh
cat << EOF | tee /opt/rotate_pcap.sh
#!/bin/sh
mv "\$1" "\${1%.incomplete}"
EOF
chmod +x /opt/rotate_pcap.sh
cat << EOF | tee /etc/systemd/system/tcpdump.service # -i is interface, -G is tick size in s, -C is filesize
[Unit]
Description="Systemd script for tcpdump"
After=network.target network-online.target
Wants=network-online.target
[Service]
User=root
WorkingDirectory=/opt/tcpdump/
ExecStart=/bin/sh -lc '/usr/bin/tcpdump -i game -C 1024 -G 180 -w "dump-%%s.pcap.incomplete" -z /opt/rotate_pcap.sh'
SuccessExitStatus=143
Restart=on-failure
RestartSec=10s
[Install]
WantedBy=multi-user.target
EOF
mkdir /opt/tcpdump
chown tcpdump:tcpdump /opt/tcpdump
cat << EOF | tee /etc/apparmor.d/local/usr.bin.tcpdump # apparmor overrides
# Allow writing any file ending in .incomplete
/**.incomplete rw,

# Allow execution of the renaming script
/usr/bin/mv ixr,
/opt/rotate_pcap.sh ixr,
EOF
sudo apparmor_parser -r /etc/apparmor.d/usr.bin.tcpdump
systemctl daemon-reload
systemctl enable tcpdump.service
systemctl start tcpdump.service
systemctl status tcpdump.service