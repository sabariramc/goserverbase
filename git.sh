export message="trace changes"
export version="v3.5.0.ddtrace"
export branch="ddtrace"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version
git push origin $branch
git push github $branch