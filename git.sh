export message="trace merge"
export version="v3.2.6.ddtrace"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version
git push origin $branch
git push github $branch