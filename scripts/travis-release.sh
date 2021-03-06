#!/bin/bash
[[ -z $TRAVIS ]] && echo "This is not Travis CI. We're not releasing." && exit
[[ -z $TRAVIS_BUILD_NUMBER || -z $TRAVIS_COMMIT ]] && echo "We're not prepared to release without a build number and commit" && exit
[[ -n $TRAVIS_PULL_REQUEST && "$TRAVIS_PULL_REQUEST" != "false" ]] && echo "No releases for pull requests" && exit

config_ssh () {
    # Be very careful not to echo anything to the logs here, none of: set -x, set -v, echo $my_private_key, env, 
    # Check the Travis logs any time you've edited this section.
    
    echo "Configuring SSH to github.com"
    [[ -z $PASSWORD ]] && echo "PASSWORD is not set!"
    touch travis_key
    chmod 600 travis_key # remove group readable, we're about to write a secret to a file, who else is on that machine right now?
    openssl aes-256-cbc -k "$PASSWORD" -d -md sha256 -a -in .id_rsa_travisci_github.priv.enc -out travis_key
    chmod 400 travis_key

    # configure ssh to github.com to use that private key
    cat >> ~/.ssh/config <<EOF
Host github.com
  IdentityFile  $PWD/travis_key
EOF

    # add github.com to the known hosts so we're not prompted on the non-interactive Travis CI
    # I copied this from my own ~/.ssh/known_host so I know it can be trusted
    known_host_github="github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
    echo $known_host_github > ~/.ssh/known_hosts 
}

cleanup () {
    echo "Cleanup"

    # reset the remote to HTTPS since git ssh won't work without the key
    echo "Resetting origin to $oldorigin"
    git remote set-url origin $oldorigin

    # delete the key
    rm -f travis_key >/dev/null 2>&1
}
trap cleanup EXIT

config_git() {
    echo "Configure git"

    # remember the origin so we can reset it during cleanup
    oldorigin=$(git remote -v | grep fetch | grep origin | awk '{print $2}')
    echo "origin was $oldorigin"

    # configure git remote "origin" to use SSH
    git remote set-url origin git@github.com:mipnw/kurl.git

    # configure user name+email
    git config --local user.email "travis@travis-ci.org"
    git config --local user.name "Travis CI"
}

git_tag_buildnumber() {
    echo "Add git tag Travis-$TRAVIS_BUILD_NUMBER"
    git tag Travis-$TRAVIS_BUILD_NUMBER $TRAVIS_COMMIT
    git push --tags
}

git_tag_incrementversion() {
    # do not increment version tag if this is a Travis build triggered by a new tag
    [[ -n $TRAVIS_TAG ]] && return

    echo "Increment version"

    # if $TRAVIS_COMMIT is already tagged, do nothing
    # we risk an infinite CI loop, and we risk parsing a tag that isn't a version
    existing_tags=`git tag --points-at $TRAVIS_COMMIT`
    [[ -n $existing_tags ]] && echo "Commit $TRAVIS_COMMIT is already tagged. Not incrementing version." && return

    # get the most recent annotated tag, it should be a version tag and not a build tag
    version=`git describe --abbrev=0`

    # try to parse the tag as a version
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

config_ssh
config_git
git_tag_incrementversion
git_tag_buildnumber
