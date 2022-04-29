export message="Added get payload"
export version="v1.4.0"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version