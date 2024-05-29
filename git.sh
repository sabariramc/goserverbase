export message="documentation update"
export version="v6.0.2"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push github $version
git push github $branch