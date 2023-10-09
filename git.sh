export message="Error handling update"
export version="v3.17.1"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch