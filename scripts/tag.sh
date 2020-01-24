#!/bin/bash

git_setup_travis() {
  [[ -z $TRAVIS ]] && return
  git config --local user.email "travis@travis-ci.org"
  git config --local user.name "Travis CI"
}

git_tag_travis_buildnumber() {
    [[ -z $TRAVIS_BUILD_NUMBER || -z $TRAVIS_COMMIT ]] && return
    test=$(git cat-file -t $TRAVIS_COMMIT)
    [[ $test != "commit" ]] && echo TRAVIS_COMMIT $TRAVIS_COMMIT is not a commit && exit 1

    git tag Travis-$TRAVIS_BUILD_NUMBER $TRAVIS_COMMIT
    git push --tags

    commit=TRAVIS_COMMIT
}

git_tag_incrementversion() {
    # do not increment version tag if this is a Travis build for a TAG
    [[ -n $TRAVIS_TAG ]] && return

    # if $commit isn't set yet, retrieve it now
    [[ -z $commit ]] && commit=`git rev-parse HEAD`

    # Only tag if the current hash isn't yet tagged
    needs_tag=`git describe --contains $commit 2>/dev/null`
    if [ -z "$needs_tag" ]; then
        # get the current version tag
        version=`git describe --abbrev=0 --tags`

        # parse the version tag
        version_bits=(${version//./ })
        major=${version_bits[0]}
        minor=${version_bits[1]}
        patch=${version_bits[2]}

        # increment the patch
        patch=$((patch+1))

        # create new tag
        next_version="$major.$minor.$patch"

        echo "Updating version tag from $version to $next_version"
        git tag $next_version
        echo "Tagged with $next_version"
        git push --tags
    else
        echo "Commit $commit is already tagged"
    fi
}

[[ -n $TRAVIS ]] && git_setup_travis && git_tag_travis_buildnumber

git_tag_incrementversion