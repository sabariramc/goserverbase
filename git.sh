export message="tracing update"
export version="v4.1.0.ddtrace"
export branch="ddtrace"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch