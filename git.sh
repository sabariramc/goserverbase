export message="Bug fixes"
export version="v1.5.1"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version