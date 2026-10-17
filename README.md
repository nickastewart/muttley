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
`MAILER=resend` does not depend on `APP_ENV`. Set `APP_BASE_URL` to the
public origin (`https://example.com`) whenever `MAILER=resend`. The
sign-in link is built only from that value. When `MAILER` is unset or
`log` and `APP_BASE_URL` is unset, the link uses `http://localhost:8080`.

## Session cookie

The `access_token` cookie and its JWT expire together after 7 days.
`COOKIE_SECURE=true` or `false` sets the Secure flag. When `COOKIE_SECURE`
is unset, the cookie is Secure only if `APP_ENV=production`. Any other
value refuses to start. The app does not read `X-Forwarded-Proto` for this.
Behind a TLS-terminating proxy, set `COOKIE_SECURE=true` or
`APP_ENV=production`.

## Test mode

Load tests that need an authenticated session set `PROFILE=test` only when
`APP_ENV` is unset or `development`:

```
APP_ENV=development
PROFILE=test
```

That registers `GET /test/login/{id}`. The route looks up that user and sets the
same `access_token` cookie as a normal sign-in, so later requests are logged in
as that user. Production must set `APP_ENV=production`. `PROFILE=test` with
`APP_ENV=production` refuses to start. Any other `APP_ENV` leaves the route
unregistered.

```
GET /test/login/1
```

Do not set `PROFILE=test` outside a local or load-test environment.

To deliver mail with [Resend](https://resend.com/docs/introduction):

```
MAILER=resend
RESEND_API_KEY=re_xxxxxxxxx
RESEND_FROM=Muttley <login@yourdomain.com>
```
