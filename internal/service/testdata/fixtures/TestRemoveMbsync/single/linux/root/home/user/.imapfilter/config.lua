
function get_pass(server, username, port)
	local status, output = pipe_from("secret-tool lookup user " .. username .. " host " .. server .. " service imap port " .. port)
assert(status == 0, "password retrieve error")
	return output
end

options.timeout = 300
options.subscribe = true

user_gmail_com = IMAP {
	server = "imap.gmail.com",
	port = 993,
	ssl = "auto",
	username = "user@gmail.com",
	password = get_pass("imap.gmail.com", "user@gmail.com", "993"),
}

results = user_gmail_com["email-archive"]:is_unseen()
results:mark_seen()
