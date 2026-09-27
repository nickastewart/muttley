***Muttley***

This project that has a http server which friends can share results from karting
and visualise them in a dashboard and leaderboard.

It is written using Golang and the Gin Framework for the backend, 
SQLite for the dastabase and HTMX for the frontend interactivity.

## Mail

Magic-link email uses the log mailer unless `MAILER=resend`. Local development
should leave `MAILER` unset so the link is printed to the server log instead of
being sent.

Load tests set the test profile:

```
PROFILE=test
```

That still creates the magic link the same way as a real sign-in: only the
token hash is stored. The link is also logged. The token route is registered
only while `PROFILE=test`. Fetch the hash with the email as a path parameter:

```
GET /test/magic-link/ada@example.com
```

```json
{"token":"<token hash>"}
```

Do not set `PROFILE=test` outside a test environment.

To deliver mail with [Resend](https://resend.com/docs/introduction):

```
MAILER=resend
RESEND_API_KEY=re_xxxxxxxxx
RESEND_FROM=Muttley <login@yourdomain.com>
```
