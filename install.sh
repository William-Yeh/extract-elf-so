#!/bin/bash
#
# Install extract-elf-so binary.
#
# USAGE:
#    install.sh  [version]  [destination directory]
#
# PREREQUISITE: awk, bash, curl.
#

VERSION=${1:-latest}
DEST_DIR=${2:-/usr/local/bin}


QUERY_URL=https://github.com/William-Yeh/extract-elf-so/releases/latest
SOURCE_URL=https://github.com/William-Yeh/extract-elf-so/releases/download/$VERSION/extract-elf-so_static_linux-amd64

TARGET_FULLPATH=$DEST_DIR/extract-elf-so


#
# error handling
#
do_error_exit () {
    echo { \"status\": $RETVAL, \"error_line\": $BASH_LINENO }
    exit
}

trap 'RETVAL=$?; echo "ERROR"; do_error_exit '  ERR
trap 'RETVAL=$?; echo "received signal to stop";  do_error_exit ' SIGQUIT SIGTERM SIGINT



get_download_url () {
    if [ "$VERSION" == "latest" ]; then
        local var=$(curl -ifsS $QUERY_URL | awk '/^Location:.+\/releases\/tag\// { print substr($2, index($2,"tag/") + 4) }')
        #echo $var

        # Remove trailing whitespace
        # @see http://stackoverflow.com/a/3352015/714426
        local latest_tag="${var%"${var##*[![:space:]]}"}"

        echo "Latest version: $latest_tag"
        SOURCE_URL=https://github.com/William-Yeh/extract-elf-so/releases/download/$latest_tag/extract-elf-so_static_linux-amd64
    fi
}



get_download_url
echo "Downloading: $SOURCE_URL"
curl -fsSL  -o $TARGET_FULLPATH  $SOURCE_URL
chmod 0755 $TARGET_FULLPATH