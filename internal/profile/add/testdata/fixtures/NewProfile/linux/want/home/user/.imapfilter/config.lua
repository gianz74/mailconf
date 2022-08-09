
function get_pass(server, username, port)
	local status, output = pipe_from("secret-tool lookup user " .. username .. " host " .. server .. " service imap port " .. port)
assert(status == 0, "password retrieve error")
	return output
end

options.timeout = 300
options.subscribe = true

jdoe_old_gmail_com = IMAP {
	server = "imap.gmail.com",
	port = 997,
	ssl = "auto",
	username = "jdoe_old@gmail.com",
	password = get_pass("imap.gmail.com", "jdoe_old@gmail.com", "997"),
}

results = jdoe_old_gmail_com["email-archive"]:is_unseen()
results:mark_seen()

jdoe_gmail_com = IMAP {
	server = "imap.gmail.com",
	port = 997,
	ssl = "auto",
	username = "jdoe@gmail.com",
	password = get_pass("imap.gmail.com", "jdoe@gmail.com", "997"),
}

results = jdoe_gmail_com["email-archive"]:is_unseen()
results:mark_seen()
