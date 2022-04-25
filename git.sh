export message="Renamed aws factory functions"
export version="v1.2.0"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version