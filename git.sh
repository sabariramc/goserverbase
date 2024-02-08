export message="tracing update"
export version="v3.19.0.ddtrace"
export branch="v3-ddtrace"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch