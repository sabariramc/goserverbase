export message="Fixed extra string empty string in the message template"
export version="v1.3.5"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version