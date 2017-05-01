corsproxy
=======

[![Build Status](https://travis-ci.org/gcochard/corsproxy.svg?branch=master)](https://travis-ci.org/gcochard/corsproxy)

This project allows you to create an App Engine instance of a cross origin
request proxy. This will let you use an App Engine instance on the free tier to
allow your super cool SPA to make requests to any site. 

This app is currently limited to only allow `GET` requests, and only for an origin
pattern defined in the `app.yaml` file.

Please copy the `app.yaml.example` to `app.yaml` and modify the
`ALLOWED_ORIGIN_REGEXP` to match your allowed origins.


Have Fun!
