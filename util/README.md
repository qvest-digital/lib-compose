# lib-compose/util

Common middleware handlers.

## NewGzipHandler

Transparently gzip the response body if the client supports it (via the Accept-Encoding header)
and the response Content-type starts with one of GzipCompressableTypes.

In difference to the most implementations found in the web,
we do the decision of compression in the Writer, when the Content-Type is determined.
