#!/bin/bash

version_properties_file=$(pwd)/deploy/version/version.properties

update_major() {
    new_version="v$(awk '{ split($0,parts,"=v"); split(parts[2],version,"."); print version[1]+1 "." version[2] "." version[3]}' $version_properties_file)";
    echo "version=${new_version}" > $version_properties_file
}

update_minor() {
    new_version="v$(awk '{ split($0,parts,"=v"); split(parts[2],version,"."); print version[1] "." version[2]+1 "." version[3]}' $version_properties_file)";
    echo "version=${new_version}" > $version_properties_file
}

update_patch() {
    new_version="v$(awk '{ split($0,parts,"=v"); split(parts[2],version,"."); print version[1] "." version[2] "." version[3]+1}' $version_properties_file)";
    echo "version=${new_version}" > $version_properties_file
}

update_kind=$1

case $update_kind in
    "--major" | "-M")
        echo "Updating major version"
        update_major
        ;;
    "--minor" | "-m")
        echo "Updating minor version"
        update_minor
        ;;
    "--patch" | "-p")
        echo "Updating patch version"
        update_patch
        ;;
    *)
        update_minor
        ;;
esac
