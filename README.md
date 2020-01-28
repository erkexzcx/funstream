# funstream

This app allows you to create your very own M3U playlist out of various bits - media files, M3U playlists, channels, streams. It also allows you to rename or exclude both groups and separate channels, override logos etc.

Supported media types:
* HLS (M3U8)
* Streams (`Content-Type: application/octet-stream`)
* Audio or Video files

**Disclaimer 1** - features, command line options and playlist format is likelly going to change in the near future once more testing and feedback is received. This also means that I will not provide binaries for a while, since everything is a subject to change.

# Roadmap

1. Investigate possibilities to proxify (multiple) EPG guides. Also ability to rewrite timestamps (e.g. +2h to each date). Add them to the same yaml file as well?
2. Performance performance performance.
3. Stability stability stability
4. Continue working on features requested by the community.
5. Docker image or SystemD service, so it properly works as a service.

# Build

For now you need to compile it yourself:
```
go build ./cmd/funstream/funstream.go
./funstream
```

# Usage

Execute binary. These command-line options are not mandatory and use only if needed:
* `-port 8989` - set custom web server's port. By default it uses `8989`.
* `-useragent "VLC/3.0.2.LibVLC/3.0.2"` - set custom user agent. By default it uses what VLC use (`VLC/3.0.2.LibVLC/3.0.2`).
* `-playlist "funstream_playlist.yaml"` - set location of your very personal funstream playlist. By default it uses `funstream_playlist.yaml` in current working directory.

# Playlist customization

See [funstream_playlist.example.yaml](https://github.com/erkexzcx/funstream/blob/master/funstream_playlist.example.yaml).

You don't need to explicitly define all values. For example, this simple one-channel `yaml` file would perfectly work:
```
channels:
  - title: ExampleTV
    url: http://example.com/path/to/stream.m3u8
    logo: http://example.com/logos/exampletv.png
    group: Example TVs
```
