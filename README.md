# Docker volume plugin for local-mapping

This plugin allows you to map local folder to volume.


## Usage


1 - Install the plugin

```
$ docker plugin install hectorqin/local-mapping

# or to enable debug
docker plugin install hectorqin/local-mapping DEBUG=1

# or to change where plugin state is stored
docker plugin install hectorqin/local-mapping state.source=<any_folder>
```

2 - Create a volume

```
$ docker volume create -d hectorqin/local-mapping -o mountpoint=/path/to/folder local-volume
local-volume
$ docker volume ls
DRIVER              VOLUME NAME
local               e1496dfe4fa27b39121e4383d1b16a0a7510f0de89f05b336aab3c0deb4dda0e
hectorqin/local-mapping:latest   local-volume

```

3 - Use the volume

```
$ docker run -it -v local-volume:<path> busybox ls <path>
```

## example


```
docker volume create -d hectorqin/local-mapping -o mountpoint=/tmp tmp
```


## LICENSE

MIT
