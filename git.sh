export message="trace mearge"
export version="v3.3.0"
export branch="ddtrace"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version
git push origin $branch
git push github $branch