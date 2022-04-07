export message="Updated get path param"
export version="v0.4.8"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version