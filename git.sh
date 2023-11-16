export message="tracing update"
export version="v3.18.8.segmentio.ddtrace"
export branch="segmentio-ddtrace"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch