export message="Added graylog"
export version="v3.4.0"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version
git push origin $branch
git push github $branch