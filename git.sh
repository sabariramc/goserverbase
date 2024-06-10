export message="fixed error not printed before wrap"
export version="v6.0.4"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push github $version
git push github $branch