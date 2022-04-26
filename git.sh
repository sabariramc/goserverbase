export message="Added AES CBC"
export version="v1.3.0"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version