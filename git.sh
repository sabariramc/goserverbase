export message="tracing update"
export version="v4.12.1.ddtrace"
export branch="ddtrace"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch