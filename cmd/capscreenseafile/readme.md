# capscreenseafile

When run in X, this command prompts the user to select an area with the mouse.

A screenshot of this area is uploaded to Seafile and a share link is copied to clipboard.

Needs [import](http://www.imagemagick.org/script/import.php) and [xclip](https://github.com/astrand/xclip) inside $PATH.

Seafile connection settings are provided through flags:

```
Usage of capscreenseafile:
  -baseurl string
    	Seafile API domain without path (default "https://your.seafile.org")
  -password string
    	Seafile password (default "password")
  -repo string
    	Seafile upload repository id (default "xxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
  -username string
    	Seafile username (default "seafile@username.org")
```
