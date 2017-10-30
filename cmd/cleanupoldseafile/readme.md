# cleanupoldseafile

Removes all files inside a seafile library that are older than a certain age.

```
Usage of cleanupoldseafile:
  -baseurl string
        Seafile API domain without path (default "https://your.seafile.org")
  -maxage duration
        A file's max age until it's cleaned up (default 336h0m0s)
  -password string
        Seafile password (default "password")
  -repo string
        Seafile upload repository id (default "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
  -username string
        Seafile username (default "your@username.org")
```
