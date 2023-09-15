SyncState *
{{ $OS := .OS}}
{{ range $Profile := .Profiles }}
IMAPAccount {{ $Profile.Name }}
Host {{ $Profile.ImapHost }}
User {{ $Profile.ImapUser }}
{{if eq $OS "linux"}}PassCmd "secret-tool lookup user {{ $Profile.ImapUser }} host {{ $Profile.ImapHost }} service imap port {{ $Profile.ImapPort }}"{{else if eq $OS "darwin"}}UseKeychain yes{{end}}
SSLType IMAPS
AuthMechs LOGIN

IMAPStore {{ $Profile.Name }}-remote
Account {{ $Profile.Name }}

MaildirStore {{ $Profile.Name }}-local
SubFolders Verbatim
Path ~/Maildir/{{ $Profile.Name }}/
Inbox ~/Maildir/{{ $Profile.Name }}/INBOX

Channel {{ $Profile.Name }}-inbox
Master :{{ $Profile.Name }}-remote:INBOX
Slave :{{ $Profile.Name }}-local:INBOX
Create Slave
Sync All
Expunge Both

Channel {{ $Profile.Name }}-trash
Master ":{{ $Profile.Name }}-remote:[Gmail]/Bin"
Slave ":{{ $Profile.Name }}-local:trash"
Create Slave
Sync All

Channel {{ $Profile.Name }}-sent
Master ":{{ $Profile.Name }}-remote:[Gmail]/Sent Mail"
Slave ":{{ $Profile.Name }}-local:sent"
Create Slave
Sync All
Expunge Both

Channel {{ $Profile.Name }}-allmail
Master ":{{ $Profile.Name }}-remote:email-archive"
Slave ":{{ $Profile.Name }}-local:email-archive"
Create Slave
Sync All
Expunge Slave

Group {{ $Profile.Name }}
Channel {{ $Profile.Name }}-inbox
Channel {{ $Profile.Name }}-trash
Channel {{ $Profile.Name }}-sent
Channel {{ $Profile.Name }}-allmail
{{ end }}