export message="Added method to get current log level"
export version="v3.10.1"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch