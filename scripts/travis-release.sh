#!/bin/bash
[[ -z $TRAVIS ]] && echo "This is not Travis CI. We're not releasing." && return
[[ -z $TRAVIS_BUILD_NUMBER || -z $TRAVIS_COMMIT ]] && echo "We're not prepared to release without a build number and commit" && return

git_setup() {
  git config --local user.email "travis@travis-ci.org"
  git config --local user.name "Travis CI"
}

git_tag_buildnumber() {
    git tag Travis-$TRAVIS_BUILD_NUMBER $TRAVIS_COMMIT
    git push --tags
}

git_tag_incrementversion() {
    # do not increment version tag if this is a Travis build triggered by a new tag
    [[ -n $TRAVIS_TAG ]] && return

    # if $TRAVIS_COMMIT is already tagged, do nothing
    # we risk an infinite CI loop, and we risk parsing a tag that isn't a version
    existing_tags=`git tag --points-at $TRAVIS_COMMIT`
    [[ -n $existing_tags ]] && echo "Commit $TRAVIS_COMMIT is already tagged. Not incrementing version." && return

    # get the most recent tag, and try to parse it as a version tag
    version=`git describe --abbrev=0 --tags`

    # parse the version tag
    version_bits=(${version//./ })

    # check that it looks like a version tag (length must be 3)
    if [[ ${#version_bits[@]} != 3 ]]; then
        echo "Most recent annotated tag $version is not a version tag"
        return
    fi

    major=${version_bits[0]}
    minor=${version_bits[1]}
    patch=${version_bits[2]}

    [[ "$(echo $major | cut -c1-1)" != "v" ]] && echo "Most recent tag $version is not a version" && return

    # increment the patch
    patch=$((patch+1))
    next_version="$major.$minor.$patch"

    # git tag
    echo "Updating version tag from $version to $next_version"
    git tag -a -m "Updating to version $next_version" $next_version $TRAVIS_COMMIT
    git push --tags
}

git_setup
git_tag_incrementversion
git_tag_buildnumber
