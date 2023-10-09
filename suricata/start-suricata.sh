#!/bin/bash

exec suricata -v -c /etc/suricata/suricata.yaml -l /eve/ -r /data/ --pcap-file-continuous --pcap-file-delete --runmode single 2>&1
