# How to execute

I think it's pretty simple, I just have to create a bunch of templates for all the different configurations that are available. Then create data structures to reference that in some sort of golang system and finally finish the thing off with a front end that maybe is auto-gen from some sort of dashboard library.

https://www.haproxy.com/documentation/dataplaneapi/community/?v=v3#overview

So change of plans, we'll just somewhat replicate the nginx proxy interface and we'll then use that to make calls to haproxy

We'll need a default config for the interface for haproxy too