export message="Test case update"
export version="v4.1.0"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch