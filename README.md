# funstream

**Disclaimer**: This software is still Work-In-Progress and many features/configuration options will change in the future.

**Funstream** allows you to build your very own HLS (M3U) playlist and proxy all its requests through this application. You can create such playlist out of separate M3U8 IPTV channels, other M3U playlist, remote/local media files. On top of that you can customize category names, channel names and logo URLs.

Some IPTV channels can be duplicates (e.g. if you combined 2 M3U playlists) and in such cases, this application works as Active-Backup mechanism, so if first used channel stops working, the second one would be used instead. :)

Advantages:
* Create single HLS (M3U) playlist out of multiple different HLs playlists, separate channels, external or internal media files.
* Proxy all the traffic through this application, effectively hiding viewer's original source IP from media source.
* Active-backup high availability for duplicating channels (e.g. if you combined 2 HLS playlists).

Disadvantages/missing features:
* No EPG support
* Based on reverse-engineering. Expect some channels/configurations not to work at all.
* No caching (if 5 viewers are watching the same IPTV channel at the same time, then IPTV channel will receive 5x more requests).

# Usage

## 1. Create playlist file

```bash
cp funstream.example.yml funstream.yml
vim funstream.yml
```

You don't need to explicitly define all fields. For example, this simple one-channel `yml` file would perfectly work:
```
channels:
  - title: ExampleTV
    url: http://example.com/path/to/stream.m3u8
    logo: http://example.com/logos/exampletv.png
    group: Example TVs
```

## 2. Build application

First, you have to download & install Golang from [here](https://golang.org/doc/install). DO NOT install Golang from the official repositories because they contain outdated version which is not working with this project.

To ensure Golang is installed successfully, test it with `go version` command. Example:
```bash
$ go version
go version go1.16.2 linux/amd64
```

Then build the application and test it:
```bash
go build -ldflags="-s -w" -o "funstream" ./cmd/funstream/main.go
./funstream -help
./funstream -playlist funstream.yml -bind 0.0.0.0:5555
```

If you decide to edit the code, you can quickly test if it works without compiling it:
```bash
go run ./cmd/funstream/main.go -help
go run ./cmd/funstream/main.go -bind 0.0.0.0:5555
```

## 4. Run application

I suggest first testing with CURL:
```bash
curl http://<ipaddr>:8888/iptv
```

You can use above URL in VLC/Kodi. :)

## 5. Installation guidelines

1. Copy/paste file `funstream.service` to `/etc/systemd/system/funstream.service`.
2. Edit `/etc/systemd/system/funstream.service` file and replace `myuser` with your non-root user. Also change paths if necessary.
3. Perform `systemctl daemon-reload`.
4. Use `systemctl <enable/disable/start/stop> funstream.service` to manage this service.

P.S. Sorry for those who are looking for binary releases or dockerfile - I will consider it when this project becomes more stable.
