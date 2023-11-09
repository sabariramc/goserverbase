export message="Http util not to relay on content length header"
export version="v3.17.2"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch