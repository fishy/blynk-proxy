# Blynk-Proxy

Blynk-Proxy is an [App Engine](https://cloud.google.com/appengine/)
app works as an HTTPS proxy for [Blynk](https://www.blynk.cc/)'s
REST API endpoints on `blynk-cloud.com`.

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

You could replace the host with `blynk-proxy.appspot.com` instead and use
HTTPS:

```
curl https://blynk-proxy.appspot.com/4ae3851817194e2596cf1b7103603ef8/pin/D8
```

`GET`, `PUT`, and `POST` HTTP methods are tested and work fine. Other HTTP
methods should also work.

## Should I use your App Engine app?

Probably not.

The only authentication on blynk API is the `auth_token` in the URL,
which means anyone who can read your HTTP requests can read and set pins on your
Blynk apps. And that's why you shouldn't use non-secure version of Blynk APIs.

Although I didn't log your requested URLs in this App Engine app intentionally,
App Engine logs all the URLs automatically and I can see them in the log.
As a result, I can see your `auth_token`(s) if I really want.

If you don't trust me (you probably shouldn't),
it's easy to sign up for an Google Cloud account yourself and deploy the code
under your account.
You won't get the nice `blynk-proxy.appspot.com` domain but everything else
should work out of the box
(except you need to change `$SELF_HOST` environment variable defined in
`app.yaml` file accordingly, but that's for a feature that is probably not used
at all).

Alternatively, you could also run this app on any (non-App-Engine) server,
you just need to set `$PORT` environment variable and run an HTTPS reverse-proxy
in front of it (e.g. [nginx](https://www.nginx.com/)).
