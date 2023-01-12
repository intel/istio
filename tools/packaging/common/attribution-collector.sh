#!/bin/bash

# Copyright (c) 2022, Intel Corporation
# SPDX-License-Identifier: BSD-3-Clause

go_licenses() {
    if ! which go > /dev/null 2>&1
    then
	echo "go not found, skipping Go licenses"
	return 1
    fi

    local dir="$1"
    local d=

    if [ -z "$dir" -o ! -d "$dir" ]
    then
	echo "No directory to look for Go licenses" >&2
	return 1
    fi

    echo "Collecting Go licenses to '${licenses}'..."

    for d in $(find "$dir" -type f -name '*.go' -printf "%h\n" | sort | uniq)
    do
	# go-licenses wants to create the license directory itself
	test -d ${licenses}-tmp && rm -rf ${licenses}-tmp

	go run github.com/google/go-licenses save "./$d" --save_path=${licenses}-tmp

	cp -nav ${licenses}-tmp/* ${licenses}
	rm -rf ${licenses}-tmp
    done

    return 0
}

is_copyleft() {
    local file="$1"
    test -r "$file" && grep -w -e MPL -e GPL -e LGPL "$file" >/dev/null && return 0

    return 1
}

deb_source() {
    local file
    local deb_install_log="$1"
    local pkg version

    if ! which apt-get > /dev/null 2>&1
    then
	ech "apt-get not found, skipping Debian source package install"
	return 1
    fi

    if [ -z "$deb_install_log" -o ! -r "$deb_install_log" ]
    then
	echo "Debian installation log file not found, skipping Debian source packages"
	return 1
    fi

    for file in /etc/apt/sources.list /etc/apt/sources.list.d/*.list
    do
	test -r "$file" || continue
	sed -n -e '/^[[:space:]]*deb / s/^[[:space:]]*deb/deb-src/p' < "$file"
    done > ${deb_source_file}

    echo "Collecting Debian source packages to '${source}'..."

    apt-get update

    cd "$source"

    grep '^Get:' "$deb_install_log" | cut -d ' ' -f 5,7 |
	while read pkg version
	do
            if ! [ -f /usr/share/doc/$pkg/copyright ]
	    then
		echo "ERROR: missing copyright file for $pkg" >&2
            fi

	    #            if matches=$(grep -w -e MPL -e GPL -e LGPL /usr/share/doc/$pkg/copyright)
	    if matches=$(is_copyleft /usr/share/doc/$pkg/copyright)
	    then
		echo "...downloading source of $pkg"
		apt-get source --download-only $pkg=$version || return 2
            else
		echo "...not downloading source of $pkg, found no copyleft license"
            fi
        done

    return 0
}

replace_with_underscores() {
    local version="$1"
    local text="$2"

    # replace strings {version} and {underscore_version}
    underscore_version="$(echo $version | sed 's/[^[:alnum:]]/_/g')"
    echo "$text" | sed -e "s/{version}/$version/g" -e "s/{underscore_version}/$underscore_version/g"
}

parse_line() {
    echo "$*" | \
	sed \
	    -e 's/[[:space:]]*\([^[:space:]]*\)[[:space:]]*=[[:space:]]*/\1=/' \
	    -e 's/\(\[[[:space:]]*\(.*\)\]\).*/\2/' \
	    -e 's/\(\[[[:space:]]*\(.*\),[[:space:]]*$\).*/\2/' \
	    -e 's/[[:space:]]%[[:space:]].*//g' \
	    -e 's/[[:space:]]*,[[:space:]]*$//'
}

envoy_licenses() {
    local file=
    local once=

    for file in $envoy_locations
    do
	test -r "$file" || continue

	if [ -z "$once" ]
	then
	       echo "Collecting Envoy licences to '$licenses'"
	       once=1
	fi
	echo "...examining Bazel file '$file'"

	# read repository_locations.bzl line by line, pick out the lines below
	# this is supposed to run from the envoy top level directory
	cat "$file" | \
	    sed -e '/urls[[:space:]]*=[[:space:]]*\[[[:space:]]*$/{N;s/\n//}' | \
	    grep -w -e "project_name" -e 'version' -e 'strip_prefix' -e 'urls' -e '),' | \
	    (
		while read line
		do
		    # the set of variables ends with '),' on its own line
		    if [ "$line" != ")," ]
		    then
			l="$(parse_line $line)"
			# the line is now '<var>=<value>', evaluate its value
			eval $l
		    else
			# last line was '),' the set of variables ends here
			# create and print variables that are being used
			echo
			echo $project_name
			echo "  version '$version'"
			archive_path="$(replace_with_underscores $version $strip_prefix)"
			echo "  archive path '$archive_path'"
			archive_url="$(replace_with_underscores $version $urls)"
			echo "  archive url '$archive_url'"

			if [ -z "$archive_url" ]
			then
			    echo "No URL for $project_name" >&2
			    failed_pkg="$failed_pkg '$project_name'"
			else
			    local copyleft=0

			    # download archive
			    mkdir -p "$download/$project_name"
			    cd "$download/$project_name"

			    wget --quiet --show-progress "$archive_url"
			    if [ $? -ne 0 ]
			    then
				echo "Could not download '$archive_url'" >&2
				failed_pkg="$failed_pkg '$project_name'"
				continue
			    fi

			    echo "  ...saved to '$download/$project_name'"

			    # extract archive
			    echo "  ...extracting"
			    case "$archive_url" in
				*.tar.gz|*.tgz)
				    tar xzf *.tar.gz 2>/dev/null
				    ;;
				*.zip)
				    unzip *.zip 2>/dev/null 1>&2
				    ;;
				*.tar.xz|*.txz)
				    tar xJf *.tar.xz 2> /dev/null 1>&2
				    ;;
				*)
				    echo "Could not extract '$archive_url'" >&2
				    failed_pkg="$failed_pkg '$project_name'"
				    continue
				    ;;
			    esac

			    # copy licence etc. information
			    mkdir -p "$licenses/$project_name"
			    cd "$download/$project_name/$archive_path"
			    found_license=0

			    for t in AUTHORS* LICENSE* LICENCE* NOTICE* CONTRIBUTORS* COPYING* COPYRIGHT* licen[cs]e*
			    do
				if [ -e "$t" ]
				then
				    cp -v $t "$licenses/$project_name"
				    found_license=1
				    is_copyleft "$t" && copyleft=1
				    if [ "$copyleft" -ne 0 -o -d "$source/$project_name" ]
				    then
					continue
				    fi

				fi
			    done
			    if [ "$found_license" -eq 0 ]
			    then
				echo "Could not find attribution files for '$project_name'" >&2
				failed_attributions="$failed_attributions '$project_name'"
			    fi

			    if [ $copyleft -ne 0 ]
			    then
				echo "...copying copyleft source"
				mkdir -p "$source/$project_name"
				cp -v "$download/$project_name"/* "$source/$project_name" 2>/dev/null
			    fi
			fi

			version=""
			strip_prefix=""
			urls=""
			archive_path=""
			archive_url=""
		    fi
		done
		echo

		if [ -n "$failed_pkg" ]
		then
		    echo "Could not get packages for$failed_pkg"
		fi
		if [ -n "$failed_attributions" ]
		then
		    echo "Could not get package attributions for$failed_attributions"
		fi
	    )
    done

    test -z "$once" && echo "Bazel build files not found, skipping"
}

usage() {
cat <<EOF
Usage: $0 [--envoy] [--go] [--debian <file>] [<file>]
	--envoy		   Search Envoy Bazel build files for software packages
	--go <dir>	   Run 'go-licenses' for Go files under director <dir>
	--debian <file>	   Get source code for Debian LGPL, GPL and MPL by
			   reading the installation log file <file>
	--prefix <dir>	   Store the atribution and source code under <dir>
EOF
}

prefix=/usr/local

get_all=1
get_envoy=0
get_go=0
get_debian=0
godir="./"

while test -n "$1"
do
    case "$1" in
	--help|h)
	    usage
	    exit 1
	    ;;
	--prefix)
	    shift
	    if [ -z "$1" ]
	    then
		echo "Error: No prefix directory given" >&2
		exit 2
	    fi

	    prefix="$1"
	    mkdir -p "$prefix" 2>/dev/null
	    if [ \! -d "$prefix" -a \! -w "$prefix" ]
	    then
		echo "Error: '$prefix' is not a writable directory" >&2
		exit 2
	    fi
	    ;;
	--envoy)
	    get_envoy=1
	    get_all=0
	    ;;
	--go)
	    shift
	    if [ -z "$1" ]
	    then
		echo "Error: No Go directory given" >&2
		exit 2
	    fi

	    godir="$1"
	    if [ \! -d "$godir" -a \! -r "$godir" ]
	    then
		echo "Error: '$godir' is not a readable directory" >&2
		exit 2
	    fi

	    get_go=1
	    get_all=0
	    ;;
	--debian)
	    shift
	    if [ -z "$1" ]
	    then
		echo "Error: No Debian installation log file found" >&2
		exit 2
	    fi
	    debian_file="$1"
	    get_debian=1
	    get_all=0
	    ;;
	*)
	    if [ -n "$debian_file" ]
	    then
		echo "Error: Debian installation log file specified twice" >&2
		exit 2
	    fi
	    debian_file="$1"
    esac
    shift
done

licenses=${LICENSE_DIR:-${prefix}/share/doc/licenses}
source=${SOURCE_DIR:-${prefix}/src}
download=${TMPDIR:-/tmp}
deb_source_file=${DEB_SOURCE_FILE:-/etc/apt/sources.list.d/deb_source.list}
envoy_locations=${ENVOY_LOCATIONS:-"bazel/repository_locations.bzl"}


for t in prefix licenses source download envoy_locations
do
    echo -e "$t: $(eval echo \$$t)"
done

mkdir -p "$licenses" "$source"

test "$get_all" == 1 -o "$get_envoy" == 1 && envoy_licenses
test "$get_all" == 1 -o "$get_go" == 1 && go_licenses "$godir"
test "$get_all" == 1 -o "$get_debian" == 1 && deb_source "$debian_file"

exit 0
