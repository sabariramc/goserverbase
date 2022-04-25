export message="Added generate id"
export version="v1.1.2"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version