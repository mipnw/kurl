#!/bin/bash
[[ -z $TRAVIS ]] && echo "This is not Travis CI. We're not releasing." && return
[[ -z $TRAVIS_BUILD_NUMBER || -z $TRAVIS_COMMIT ]] && echo "We're not prepared to release without a build number and commit" && return
[[ -n $TRAVIS_PULL_REQUEST ]] && echo "No releases for pull requests" && return

config_ssh () {
    # Be very careful not to echo anything to the logs here, no set -x, or set -v, or echo $my_private_key. 
    # Check the Travis logs any time you've edited this section.
    
    echo "Configuring SSH to github.com"
    touch travis_key
    chmod 600 travis_key # remove group readable, we're about to write a secret to a file, who else is on that machine right now?
    openssl aes-256-cbc -k "$TRAVIS_KEY_PASSWORD" -d -md sha256 -a -in .id_rsa_travisci_github.priv.enc -out travis_key
    chmod 400 travis_key

    # configure ssh to github.com to use that private key
    cat >> ~/.ssh/config <<EOF
Host github.com
  IdentityFile  $(pwd)/travis_key
EOF

    # add github.com to the known hosts so we're not prompted on the non-interactive Travis CI
    # I copied this from my own ~/.ssh/known_host so I know it can be trusted
    known_host_github="github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
    echo $known_host_github > ~/.ssh/known_hosts 
}

config_git() {
    # configure git remote "origin" to use SSH
    git remote set-url origin ssh://git@github.com:mipnw/kurl.git

    # configure user name+email
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

config_ssh
config_git
git_tag_incrementversion
git_tag_buildnumber
