export message="Refactored AES"
export version="v1.3.4"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version