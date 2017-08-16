#!/usr/bin/env bash

set -e

repo_path="go.zenithar.org/password"

version=$( cat version/VERSION )
revision=$( git rev-parse --short HEAD 2> /dev/null || echo 'unknown' )
branch=$( git rev-parse --abbrev-ref HEAD 2> /dev/null || echo 'unknown' )
host=$( hostname -f )
build_date=$( date +%Y%m%d-%H:%M:%S )
go_version=$( go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/' )

if [ "$(go env GOOS)" = "windows" ]; then
	ext=".exe"
fi

ldflags="
  -X ${repo_path}/version.Version=${version}
  -X ${repo_path}/version.Revision=${revision}
  -X ${repo_path}/version.Branch=${branch}
  -X ${repo_path}/version.BuildUser=${USER}@${host}
  -X ${repo_path}/version.BuildDate=${build_date}
  -X ${repo_path}/version.GoVersion=${go_version}
  -s"

echo " >   server"
go build -buildmode=pie -ldflags "${ldflags}" -o bin/password_server${ext} ${repo_path}

exit 0
