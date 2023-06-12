export message="Trace update"
export branch="ddtrace"
export version="v3.2.3.ddtrace"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version
git push origin $branch
git push github $branch