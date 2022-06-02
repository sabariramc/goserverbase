export message="Added kafka"
export version="v1.5.0"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version