export message="tracing update"
export version="v3.17.2.ddtrace"
export branch="ddtrace"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch