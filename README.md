***Muttley***

This project that has a http server which friends can share results from karting
and visualise them in a dashboard and leaderboard.

It is written using Golang and the Gin Framework for the backend, 
SQLite for the dastabase and HTMX for the frontend interactivity.

## Mail

Magic-link email uses the log mailer when `APP_ENV=development` and `MAILER`
is unset or `log`. Local development should set `APP_ENV=development` and leave
`MAILER` unset so the link is printed to the server log instead of being sent.
Any other `APP_ENV`, including unset, refuses to start with the log mailer.
`MAILER=resend` does not depend on `APP_ENV`.

## Test mode

Load tests that need an authenticated session set:

```
PROFILE=test
```

That registers `GET /test/login/{id}`. The route looks up that user and sets the
same `access_token` cookie as a normal sign-in, so later requests are logged in
as that user. Any other profile leaves the route unregistered.

```
GET /test/login/1
```

Do not set `PROFILE=test` outside a test environment.

To deliver mail with [Resend](https://resend.com/docs/introduction):

```
MAILER=resend
RESEND_API_KEY=re_xxxxxxxxx
RESEND_FROM=Muttley <login@yourdomain.com>
```
