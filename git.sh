export message="Consumer channel closure"
export version="v3.7.2"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version
git push origin $branch
git push github $branch