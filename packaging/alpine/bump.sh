#!/usr/bin/env bash
#
# Prepare the Alpine aports update for a gomplate release.
#
# usage: packaging/alpine/bump.sh <version> [path-to-aports-checkout]
#
# Needs gomplate and git. Renders APKBUILD.tmpl (using the release's published
# source checksum) into community/gomplate/APKBUILD on a new branch of a local
# alpine/aports checkout, and commits it. Pushing and opening the merge request
# are left to you - see the instructions printed at the end.
#
# The aports checkout defaults to ~/go/src/gitlab.alpinelinux.org/alpine/aports
# and is expected to have "origin" pointing at alpine/aports.

set -euo pipefail

version="${1:?usage: $0 <version> [path-to-aports-checkout]}"
version="${version#v}"
aports="${2:-$HOME/go/src/gitlab.alpinelinux.org/alpine/aports}"
here="$(cd "$(dirname "$0")" && pwd)"
branch="upgrade-gomplate-aport-${version}"

# GitLab refuses web/git-over-HTTPS from CI runners (bot filter), so this is
# meant to be run by hand
git -C "$aports" remote get-url fork >/dev/null 2>&1 ||
	git -C "$aports" remote add fork git@gitlab.alpinelinux.org:hairyhenderson/aports.git

git -C "$aports" checkout -q master
git -C "$aports" pull -q --ff-only origin master
git -C "$aports" checkout -q -b "$branch"

VERSION="$version" gomplate \
	-c "sha512=https://github.com/hairyhenderson/gomplate/releases/download/v${version}/v${version}.sha512?type=text/plain" \
	-f "$here/APKBUILD.tmpl" \
	-o "$aports/community/gomplate/APKBUILD"

git -C "$aports" --no-pager diff
git -C "$aports" commit -sqam "community/gomplate: upgrade to ${version}"

cat <<EOF

Committed on branch ${branch} in ${aports}. Next:

  cd ${aports}
  git push fork HEAD        # prints a link to create the merge request

In the merge request, tick "Allow commits from members who can merge to the
target branch", and add this to the description:

  https://github.com/hairyhenderson/gomplate/releases/tag/v${version}
EOF
