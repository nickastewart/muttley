***Muttley***

This project that has a http server which friends can share results from karting
and visualise them in a dashboard and leaderboard.

It is written using Golang and the Gin Framework for the backend, 
SQLite for the dastabase and HTMX for the frontend interactivity.

## Mail

Magic-link email uses the log mailer unless `MAILER=resend`. Local development
should leave `MAILER` unset so the link is printed to the server log instead of
being sent.

Load tests can set `MAILER=test`. That still creates the magic link the same
way as a real sign-in, logs it, and saves the raw token on that database row.
Fetch it with:

```
GET /test/magic-link?email=ada@example.com
```

```json
{"token":"..."}
```

Do not enable `MAILER=test` outside a test environment. The endpoint returns a
token that signs that user in.

To deliver mail with [Resend](https://resend.com/docs/introduction):

```
MAILER=resend
RESEND_API_KEY=re_xxxxxxxxx
RESEND_FROM=Muttley <login@yourdomain.com>
```
