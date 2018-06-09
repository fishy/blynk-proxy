# Blynk-Proxy

Blynk-Proxy is a [Heroku](https://heroku.com) app works as an HTTPS proxy for
[Blynk](https://www.blynk.cc/)'s REST API endpoints on `blynk-cloud.com`.

## Why?

Because as of the time of writing,
`blynk-cloud.com` only provides a self-signed cert for HTTPS.

With this proxy, you can access the same APIs through HTTPS requests with proper
certs.

## How to Use

For a Blynk REST API you would normally access through `blynk-cloud.com`:

```
curl http://blynk-cloud.com/4ae3851817194e2596cf1b7103603ef8/pin/D8
```

Or

```
curl https://blynk-cloud.com/4ae3851817194e2596cf1b7103603ef8/pin/D8
```

You could replace the host with `blynk-proxy.herokuapp.com` instead and use
HTTPS:

```
curl https://blynk-proxy.herokuapp.com/4ae3851817194e2596cf1b7103603ef8/pin/D8
```

`GET`, `PUT`, and `POST` HTTP methods are tested and work fine. Other HTTP
methods should also work.

## Should I use your Heroku app?

Probably not.

The only authentication on blynk API is the `auth_token` in the URL,
which means anyone who can read your HTTP requests can read and set pins on your
Blynk apps. And that's why you shouldn't use non-secure version of Blynk APIs.

Although I didn't log your requested URLs in this Heroku app intentionally,
Heroku logs all the URLs automatically and I can see them in the log.
I also do not want to disable the log altogether (if I can) because I still need
the ability to debug if things go wrong.
As a result, I can see your `auth_token`(s) if I really want.

Also if this gets popular and enough people uses my Heroku app,
it might no longer fit in their free tier and I might need to pay Heroku for
this app.

If you don't trust me (you probably shouldn't),
it's easy to sign up for a Heroku account yourself and deploy the code under
your account.
You won't get the nice `blynk-proxy.herokuapp.com` domain but everything else
should work out of the box
(except you need to change `selfHost` in the code accordingly,
but that's for a feature that is probably not used at all).

Alternatively, you could also run this app on any (non-Heroku) server,
you just need to set `$PORT` environment variable and run an HTTPS reverse-proxy
in front of it (e.g. [nginx](https://www.nginx.com/)).
You probably want to remove the Heroku import in the code if you are not running
it on Heroku.
