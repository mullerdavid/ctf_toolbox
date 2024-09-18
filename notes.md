# Docker 

'''bash
docker compose up -d --build
docker compose up -d --build --force-recreate
docker compose up -d --no-deps --build --force-recreate arkime
docker compose exec pcapmover sh -c "printf '' | tee -a /data/*.pcap"
'''

# WSL elasticsearch map count
'''bash
wsl -d docker-desktop
sudo sysctl -w vm.max_map_count=262144
'''

# tshark
'''bash
tshark -G elastic-mapping --elastic-mapping-filter http,tcp,udp,ip | jq -f mapping-transform.jq > mapping.json
curl -X PUT "localhost:9200/_index_template/packets_template" -H 'Content-Type: application/json' --data-binary @mapping.json
tshark -T ek -J "http tcp udp ip" -x -r ./dump-1721489046.pcap | go run tshark-to-elastic.go "http://localhost:9200/packets_template/_bulk"
'''

# Kibana DSL
'''json
{
  "regexp": {
    "layers.tcp.tcp_tcp_payload_raw": {
      "case_insensitive": true,
      "value": "[\u0000-\u001f]{32} .*0123456789:;\\<=\\>\\?@[\u0000-A]BCDEFGHIJK[LM]{2}NOPQRSTUVWXYZ.*"
    }
  }
}
'''

# TODO
 - .env memory limits
 - check http gzip
 - suricata disable signatures test