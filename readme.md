# Go tcp copy

Buy me a cup of coffee for $3

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/M4M54KKIF)

This application can forward tcp connection to upstream, copy to more than one upstream.

## build

```
CGO_ENABLED=0 go build -v -a -ldflags ' -s -w  -extldflags "-static"' .
```

## use

```
  -l string
    	listen host:port
  -r string
    	remote host:port, if mutiple, split with ,

```

Example use, forward local port to remote server port

```
./gotcpcopy -l :3306 -r 10.1.23.43:3316,10.1.23.45:1176,10.1.17.12:3189
```


## License

```
    Copyright (C) 2000-2022 cnmade

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
```
