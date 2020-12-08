# secret-gift-emailer

Secret Gift Emailer Thingy

- Send gmail with Go https://github.com/gaurangmacharya/pepithon/blob/master/send-email-via-gmail-smtp-server-using-go.go

## How to

**Prereqs**:

1. Disable 2-factor auth on your gmail (RE-ENABLE WHEN DONE with this app)
2. Allow less secure apps https://myaccount.google.com/u/0/lesssecureapps (DISABLE WHEN DONE!)
3. Install deps `go get`

**Run it**:

1. `GMAIL_EMAIL="my-email@gmail.com" CSV=test.csv go run main.go`
2. Profit!

> Tip set `DEBUG=true` to do a dry run and show email config output
