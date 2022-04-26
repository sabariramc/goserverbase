export message="Updated error handling of AES CBC"
export version="v1.3.1"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version