export message="Looging update for aws services"
export version="v1.2.1"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version