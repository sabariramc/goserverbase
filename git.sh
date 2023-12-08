export message="Master merge"
export version="v4.0.1"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch