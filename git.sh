export message="Fixed logger service name"
export version="v1.3.6"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version