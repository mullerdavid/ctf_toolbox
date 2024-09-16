{
    "settings": ( .settings * {"index.mapping.ignore_malformed": true, "index.mapping.coerce": true} ), 
    "mappings": .mappings, 
    "index_patterns": ["packets*"] 
}
| walk(if type == "object" and .Name == "type" and .Value == "boolean" then .Value = "short" else . end)