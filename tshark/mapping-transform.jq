{
    "template": {
        "settings": ( .settings * {
            "index.mapping.ignore_malformed": true, 
            "index.mapping.coerce": true
        } ), 
        "mappings": ( .mappings * {
            "properties": {
                "layers": {
                    "properties": {
                        "http": {
                            "properties": {
                                "http_http_request_line_raw": {
                                    "type": "text"
                                },
                                "http_http_response_line_raw": {
                                    "type": "text"
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