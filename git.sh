export message="Renamed kafka config variable"
export version="v3.17.0"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch