export message="New version"
export version="v4.0.0"
export branch="v4"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch