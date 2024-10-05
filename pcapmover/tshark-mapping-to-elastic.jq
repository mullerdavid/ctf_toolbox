{
    "template": {
        "settings": ( .settings * {
            "index.mapping.ignore_malformed": true, 
            "index.mapping.coerce": true
        } ), 
        "mappings": ( .mappings * {
            "properties": {
                "flag": {
                    "type": "boolean"
                },
                "file": {
                    "type": "text"
                },
                "layers": {
                    "properties": {
                        "udp": {
                            "properties": {
                                "udp_udp_payload_raw": {
                                    "type": "wildcard"
                                }
                            }
                        },
                        "tcp": {
                            "properties": {
                                "tcp_tcp_payload_raw": {
                                    "type": "wildcard"
                                }
                            }
                        },
                        "http": {
                            "properties": {
                                "http_http_request_line": {
                                    "type": "text"
                                },
                                "http_http_request_line_raw": {
                                    "type": "wildcard"
                                },
                                "http_http_response_line": {
                                    "type": "text"
                                },
                                "http_http_response_line_raw": {
                                    "type": "wildcard"
                                }
                            }
                        }
                    }
                }
            }
        } ), 
    },
    "index_patterns": ["packets*"] 
}
| walk(if type == "object" and .Name == "type" and .Value == "boolean" then .Value = "short" else . end)