export message="Test Update"
export version="v1.4.1"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version