local ENV_FLAG = os.getenv("FLAG_REGEX")
local REGEX_FLAG = ArkimeData.pcre_create(ENV_FLAG)

print("Flag LUA plugin loaded with FLAG_REGEX:", ENV_FLAG)

local function parser_flag(session, str, direction)
    data = ArkimeData.new(str)
    local matched = data:pcre_match(REGEX_FLAG)
    if matched then
        session:add_tag("FLAG")
        if direction == 0 then -- request
            session:add_tag("FLAG_IN")
        else
            session:add_tag("FLAG_OUT")
        end
        
    end
end

function register_parser(session, str, direction)
    session:register_parser(parser_flag)
end

ArkimeSession.register_tcp_classifier("flag", 0, "", "register_parser")
ArkimeSession.register_udp_classifier("flag", 0, "", "register_parser")