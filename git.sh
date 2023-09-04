export message="Logger reference changes"
export version="v3.15.1"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch