export message="Version change "
export version="v3.6.1"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version
git push origin $branch
git push github $branch